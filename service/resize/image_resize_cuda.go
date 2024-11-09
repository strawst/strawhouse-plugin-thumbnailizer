//go:build cuda

package resize

/*
#cgo LDFLAGS: -L. -lbicubic -lcuda -lcudart
#include <cuda_runtime.h>
#include <stdint.h>

extern void bicubicInterpolation(uint8_t* src, int srcWidth, int srcHeight,
                                  int targetWidth, int targetHeight, uint8_t* dst);
*/
import "C"
import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"unsafe"

	"github.com/bsthun/gut"
)

func ResizeImage(src image.Image, targetPixels int) ([]byte, *gut.ErrorInstance) {
	// Calculate the target dimensions while maintaining the aspect ratio
	aspectRatio := float64(src.Bounds().Dx()) / float64(src.Bounds().Dy())
	targetWidth := int(math.Sqrt(float64(targetPixels) * aspectRatio))
	targetHeight := int(float64(targetPixels) / float64(targetWidth))

	// Create a new image with the target dimensions
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Prepare source image data for CUDA
	srcWidth := src.Bounds().Dx()
	srcHeight := src.Bounds().Dy()
	srcData := make([]uint8, srcWidth*srcHeight*4) // RGBA

	for y := 0; y < srcHeight; y++ {
		for x := 0; x < srcWidth; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			srcData[(y*srcWidth+x)*4+0] = uint8(r >> 8)
			srcData[(y*srcWidth+x)*4+1] = uint8(g >> 8)
			srcData[(y*srcWidth+x)*4+2] = uint8(b >> 8)
			srcData[(y*srcWidth+x)*4+3] = uint8(a >> 8)
		}
	}

	// Allocate memory on the device
	var srcDevice, dstDevice *C.uint8_t
	C.cudaMalloc(unsafe.Pointer(&srcDevice), C.size_t(len(srcData)))
	C.cudaMalloc(unsafe.Pointer(&dstDevice), C.size_t(targetWidth*targetHeight*4))

	// Copy data to the device
	C.cudaMemcpy(srcDevice, unsafe.Pointer(&srcData[0]), C.size_t(len(srcData)), C.cudaMemcpyHostToDevice)

	// Define CUDA grid and block sizes
	blockSize := 16
	gridWidth := (targetWidth + blockSize - 1) / blockSize
	gridHeight := (targetHeight + blockSize - 1) / blockSize
	dimGrid := C.dim3{X: C.uint(gridWidth), Y: C.uint(gridHeight), Z: 1}
	dimBlock := C.dim3{X: C.uint(blockSize), Y: C.uint(blockSize), Z: 1}

	// Launch the CUDA kernel
	C.bicubicInterpolation(srcDevice, C.int(srcWidth), C.int(srcHeight),
		C.int(targetWidth), C.int(targetHeight), dstDevice)

	// Copy back the result from device to host
	dstData := make([]uint8, targetWidth*targetHeight*4)
	C.cudaMemcpy(unsafe.Pointer(&dstData[0]), dstDevice, C.size_t(len(dstData)), C.cudaMemcpyDeviceToHost)

	// Set the resized image data
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			dst.Set(x, y, color.RGBA{
				R: dstData[(y*targetWidth+x)*4+0],
				G: dstData[(y*targetWidth+x)*4+1],
				B: dstData[(y*targetWidth+x)*4+2],
				A: dstData[(y*targetWidth+x)*4+3],
			})
		}
	}

	// Clean up device memory
	C.cudaFree(srcDevice)
	C.cudaFree(dstDevice)

	// Encode the resized image to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, nil); err != nil {
		return nil, gut.Err(false, "error encoding image", err)
	}

	return buf.Bytes(), nil
}
