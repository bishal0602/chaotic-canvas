package utils

import (
	"image/color"
	"math/rand"
)

func RandomRGBA() color.RGBA {
	return color.RGBA{
		R: uint8(rand.Intn(256)),
		G: uint8(rand.Intn(256)),
		B: uint8(rand.Intn(256)),
		A: uint8(rand.Intn(206) + 50), // Alpha between 50-255 for semi-transparency
	}
}
