package mathutil

import "math/rand"

// RandomBetween returns a random integer between a and b (inclusive).
func RandomBetween(a, b int) int {
	if a > b {
		a, b = b, a // Swap if a > b to avoid errors
	}
	return rand.Intn(b-a+1) + a
}
