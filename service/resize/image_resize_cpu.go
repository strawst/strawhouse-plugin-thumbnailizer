//go:build !cuda

package resize

import (
	"bytes"
	"github.com/bsthun/gut"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"
)

const (
	chunkSize = 16
)

func ResizeImage(src image.Image, targetPixels int, quality float32) ([]byte, *gut.ErrorInstance) {
	// Calculate the target dimensions while maintaining the aspect ratio
	bounds := src.Bounds()
	aspectRatio := float64(bounds.Dx()) / float64(bounds.Dy())
	targetWidth := int(math.Sqrt(float64(targetPixels) * aspectRatio))
	targetHeight := int(float64(targetPixels) / float64(targetWidth))

	// Declare variables
	var dst image.Image
	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	chunks := make(chan chunk, workers)

	// Handle target larger than source
	if targetWidth >= bounds.Dx() {
		dst = src
		goto encode
	}

	// Create a new image with the target dimensions
	dst = image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Calculate number of worker goroutines based on CPU cores

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range chunks {
				processChunk(src, dst.(*image.RGBA), c, bounds, targetWidth, targetHeight)
			}
		}()
	}

	// Distribute work in chunks
	for startY := 0; startY < targetHeight; startY += chunkSize {
		endY := minimum(startY+chunkSize, targetHeight)
		for startX := 0; startX < targetWidth; startX += chunkSize {
			endX := minimum(startX+chunkSize, targetWidth)
			chunks <- chunk{startX, endX, startY, endY}
		}
	}
	close(chunks)
	wg.Wait()

	// Encode image
encode:
	var buf bytes.Buffer
	options, _ := encoder.NewLossyEncoderOptions(encoder.PresetPhoto, quality)
	options.Method = 6
	if err := webp.Encode(&buf, dst, options); err != nil {
		return nil, gut.Err(false, "error encoding image", err)
	}

	return buf.Bytes(), nil
}

type chunk struct {
	startX, endX, startY, endY int
}

func processChunk(src image.Image, dst *image.RGBA, c chunk, bounds image.Rectangle, targetWidth, targetHeight int) {
	scaleX := float64(bounds.Dx()) / float64(targetWidth)
	scaleY := float64(bounds.Dy()) / float64(targetHeight)

	for y := c.startY; y < c.endY; y++ {
		for x := c.startX; x < c.endX; x++ {
			srcX := float64(x) * scaleX
			srcY := float64(y) * scaleY
			col := bicubicInterpolation(src, srcX, srcY)
			dst.Set(x, y, optimizeColor(col))
		}
	}
}

func optimizeColor(c color.Color) color.Color {
	r, g, b, a := c.RGBA()

	// Convert to YCbCr color space
	y := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	cb := 128 - 0.168736*float64(r) - 0.331264*float64(g) + 0.5*float64(b)
	cr := 128 + 0.5*float64(r) - 0.418688*float64(g) - 0.081312*float64(b)

	// quantize color depth
	y = math.Round(y/256) * 256
	cb = math.Round(cb/256) * 256
	cr = math.Round(cr/256) * 256

	// Convert back to RGB
	r = uint32(y + 1.402*(cr-128))
	g = uint32(y - 0.344136*(cb-128) - 0.714136*(cr-128))
	b = uint32(y + 1.772*(cb-128))

	return color.RGBA{
		R: uint8(clamp(int(r>>8), 0, 255)),
		G: uint8(clamp(int(g>>8), 0, 255)),
		B: uint8(clamp(int(b>>8), 0, 255)),
		A: uint8(clamp(int(a>>8), 0, 255)),
	}
}

func bicubicInterpolation(img image.Image, x, y float64) color.Color {
	x1 := int(x)
	y1 := int(y)
	bounds := img.Bounds()

	// Get the 16 surrounding pixels
	var pixels [4][4]color.Color
	for i := -1; i <= 2; i++ {
		for j := -1; j <= 2; j++ {
			px := clamp(x1+i, 0, bounds.Dx()-1)
			py := clamp(y1+j, 0, bounds.Dy()-1)
			pixels[i+1][j+1] = img.At(px, py)
		}
	}

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

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
