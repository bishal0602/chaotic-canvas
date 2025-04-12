package imageio

import (
	"image"
	"image/color"
	"testing"
)

// createTestImage returns a new RGBA test image of the specified width and height with a specified color.
func createTestImage(width, height int, col color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}
	return img
}

func TestResize_NoResize(t *testing.T) {
	// For an image smaller than the max dimension, Resize should return the same image.
	const maxDim = 540
	origWidth, origHeight := 300, 400
	origImage := createTestImage(origWidth, origHeight, color.RGBA{R: 100, G: 150, B: 200, A: 255})

	resized := Resize(origImage, maxDim)
	bounds := resized.Bounds()
	if bounds.Dx() != origWidth || bounds.Dy() != origHeight {
		t.Errorf("Expected dimensions %dx%d, got %dx%d", origWidth, origHeight, bounds.Dx(), bounds.Dy())
	}
}

func TestResize_RescaleWidth(t *testing.T) {
	// Test when width is the dominant dimension.
	const maxDim = 540
	origWidth, origHeight := 800, 600
	origImage := createTestImage(origWidth, origHeight, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	resized := Resize(origImage, maxDim)
	bounds := resized.Bounds()

	// When width > height, new width should equal maxDim.
	expectedWidth := maxDim
	// Aspect ratio preservation: newHeight = origHeight * maxDim / origWidth.
	expectedHeight := int(float64(origHeight) * float64(maxDim) / float64(origWidth))

	if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
		t.Errorf("Expected dimensions %dx%d, got %dx%d", expectedWidth, expectedHeight, bounds.Dx(), bounds.Dy())
	}
}

func TestResize_RescaleHeight(t *testing.T) {
	// Test when height is the dominant dimension.
	const maxDim = 540
	origWidth, origHeight := 400, 800
	origImage := createTestImage(origWidth, origHeight, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	resized := Resize(origImage, maxDim)
	bounds := resized.Bounds()

	// When height > width, new height should equal maxDim.
	expectedHeight := maxDim
	expectedWidth := int(float64(origWidth) * float64(maxDim) / float64(origHeight))

	if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
		t.Errorf("Expected dimensions %dx%d, got %dx%d", expectedWidth, expectedHeight, bounds.Dx(), bounds.Dy())
	}
}
