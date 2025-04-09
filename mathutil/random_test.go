package mathutil

import (
	"testing"
)

func TestRandomBetween(t *testing.T) {
	min, max := -10, 20
	for i := 0; i < 100; i++ {
		val := RandomBetween(min, max)
		if val < min || val > max {
			t.Errorf("RandomBetween(%d,%d) produced %d; out of range", min, max, val)
		}
		// Test swap: if a > b it should swap
		val = RandomBetween(max, min)
		if val < min || val > max {
			t.Errorf("RandomBetween(%d,%d) produced %d; out of expected range", max, min, val)
		}
	}
}
