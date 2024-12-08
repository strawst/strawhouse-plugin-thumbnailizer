//go:build !cuda

package resize

import (
	"bytes"
	"github.com/bsthun/gut"
	"image"
	"image/jpeg"
	"math"
)

func ResizeImage(src image.Image, targetPixels int) ([]byte, *gut.ErrorInstance) {
	// Calculate the target dimensions while maintaining the aspect ratio
	aspectRatio := float64(src.Bounds().Dx()) / float64(src.Bounds().Dy())
	targetWidth := int(math.Sqrt(float64(targetPixels) * aspectRatio))
	targetHeight := int(float64(targetPixels) / float64(targetWidth))

	// Create a new image with the target dimensions
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Resize the image using nearest-neighbor interpolation
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			// Calculate the nearest neighbor's coordinates in the source image
			srcX := int(float64(x) * float64(src.Bounds().Dx()) / float64(targetWidth))
			srcY := int(float64(y) * float64(src.Bounds().Dy()) / float64(targetHeight))

			// Get the color of the nearest neighbor and set it to the destination pixel
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	// Encode the resized image to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, nil); err != nil {
		return nil, gut.Err(false, "error encoding image", err)
	}

	return buf.Bytes(), nil
}
