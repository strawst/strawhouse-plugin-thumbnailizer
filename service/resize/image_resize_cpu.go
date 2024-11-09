//go:build !cuda

package resize

import (
	"bytes"
	"github.com/bsthun/gut"
	"image"
	"image/color"
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

	// Resize the image using bicubic interpolation
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := float64(x) * float64(src.Bounds().Dx()) / float64(targetWidth)
			srcY := float64(y) * float64(src.Bounds().Dy()) / float64(targetHeight)
			dst.Set(x, y, bicubicInterpolation(src, srcX, srcY))
		}
	}

	// Encode the resized image to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, nil); err != nil {
		return nil, gut.Err(false, "error encoding image", err)
	}

	return buf.Bytes(), nil
}

func bicubicInterpolation(img image.Image, x, y float64) color.Color {
	x1 := int(x)
	y1 := int(y)

	// Get the 16 surrounding pixels
	var pixels [4][4]color.Color
	for i := -1; i <= 2; i++ {
		for j := -1; j <= 2; j++ {
			px := clamp(x1+i, 0, img.Bounds().Dx()-1)
			py := clamp(y1+j, 0, img.Bounds().Dy()-1)
			pixels[i+1][j+1] = img.At(px, py)
		}
	}

	// Perform bicubic interpolation
	return bicubic(pixels, x-float64(x1), y-float64(y1))
}

func bicubic(pixels [4][4]color.Color, dx, dy float64) color.Color {
	var r, g, b, a float64
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			weight := cubicWeight(float64(i)-1-dx) * cubicWeight(float64(j)-1-dy)
			ri, gi, bi, ai := pixels[i][j].RGBA()
			r += weight * float64(ri)
			g += weight * float64(gi)
			b += weight * float64(bi)
			a += weight * float64(ai)
		}
	}
	return color.RGBA{
		R: uint8(clamp(int(r/256), 0, 255)),
		G: uint8(clamp(int(g/256), 0, 255)),
		B: uint8(clamp(int(b/256), 0, 255)),
		A: uint8(clamp(int(a/256), 0, 255)),
	}
}

func cubicWeight(t float64) float64 {
	t = math.Abs(t)
	if t <= 1 {
		return 1.5*t*t*t - 2.5*t*t + 1
	} else if t <= 2 {
		return -0.5*t*t*t + 2.5*t*t - 4*t + 2
	}
	return 0
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
