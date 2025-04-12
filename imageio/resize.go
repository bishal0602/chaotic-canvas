package imageio

import (
	"image"
	"image/color"
	"math"

	"github.com/bishal0602/chaotic-canvas/mathutil"
)

// Resize limits the maximum width and height to maxDim while preserving the aspect ratio.
func Resize(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// No resizing needed if both dimensions are within the limit.
	if width <= maxDim && height <= maxDim {
		return img
	}

	// Calculate new dimensions while preserving the aspect ratio.
	var newWidth, newHeight int
	if width > height {
		newWidth = maxDim
		newHeight = int(float64(height) * float64(maxDim) / float64(width))
	} else {
		newHeight = maxDim
		newWidth = int(float64(width) * float64(maxDim) / float64(height))
	}

	// Use a high-quality resampling algorithm.
	resizedImg := resizeBilinear(img, newWidth, newHeight)
	return resizedImg
}

// resizeBilinear resizes the input image to the given width and height using bilinear interpolation.
func resizeBilinear(src image.Image, newWidth, newHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	xScale := float64(srcWidth) / float64(newWidth)
	yScale := float64(srcHeight) / float64(newHeight)
	xOffset := 0.5*xScale - 0.5
	yOffset := 0.5*yScale - 0.5

	for y := 0; y < newHeight; y++ {
		srcY := float64(y)*yScale + yOffset
		y0 := int(math.Floor(srcY))
		y1 := y0 + 1
		v := srcY - float64(y0)
		// Clamp y0 and y1 to the source bounds.
		y0 = mathutil.Max(0, y0)
		y1 = mathutil.Min(srcHeight-1, y1)

		for x := 0; x < newWidth; x++ {
			srcX := float64(x)*xScale + xOffset
			x0 := int(math.Floor(srcX))
			x1 := x0 + 1
			u := srcX - float64(x0)
			x0 = mathutil.Max(0, x0)
			x1 = mathutil.Min(srcWidth-1, x1)

			// Get the four surrounding pixels.
			c00 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y0)
			c01 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y0)
			c10 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y1)
			c11 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y1)

			// Convert the colors to floating point values on a 0-255 scale.
			r00, g00, b00, a00 := colorToFloat(c00)
			r01, g01, b01, a01 := colorToFloat(c01)
			r10, g10, b10, a10 := colorToFloat(c10)
			r11, g11, b11, a11 := colorToFloat(c11)

			// Interpolate each channel.
			r := bilinear(r00, r01, r10, r11, u, v)
			g := bilinear(g00, g01, g10, g11, u, v)
			b := bilinear(b00, b01, b10, b11, u, v)
			a := bilinear(a00, a01, a10, a11, u, v)

			// Set the new pixel in the destination image.
			dst.Set(x, y, color.NRGBA{
				R: uint8(mathutil.Clamp(r, 0, 255)),
				G: uint8(mathutil.Clamp(g, 0, 255)),
				B: uint8(mathutil.Clamp(b, 0, 255)),
				A: uint8(mathutil.Clamp(a, 0, 255)),
			})
		}
	}
	return dst
}

// colorToFloat converts a color.Color to its RGBA components as float64 values in 0-255 range.
func colorToFloat(c color.Color) (float64, float64, float64, float64) {
	r, g, b, a := c.RGBA()
	// Division by 257 correctly scales a uint16 (0-65535) to a float64 (0.0-255.0)
	return float64(r) / 257.0, float64(g) / 257.0, float64(b) / 257.0, float64(a) / 257.0
}

// bilinear performs bilinear interpolation given the four corner values and interpolation factors u and v.
func bilinear(c00, c01, c10, c11, u, v float64) float64 {
	return (1-u)*(1-v)*c00 +
		u*(1-v)*c01 +
		(1-u)*v*c10 +
		u*v*c11
}
