package utils

import "testing"

func TestFloorPowerOfTen(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 1},                   // 0 default to 1
		{5, 1},                   // 1-9 → 1
		{23, 10},                 // 10-99 → 10
		{150, 100},               // 100-999 → 100
		{999, 100},               // 100-999 → 100
		{1001, 1000},             // 1000-9999 → 1000
		{45678, 10000},           // 10,000-99,999 → 10,000
		{987654, 100000},         // 100,000-999,999 → 100,000
		{987654321, 100000000},   // 100,000,000-999,999,999 → 100,000,000
		{9876543210, 1000000000}, // 1,000,000,000-9,999,999,999 → 1,000,000,000
		{-10, 1},                 // Negative numbers default to 1
	}

	for _, tt := range tests {
		result := FloorPowerOfTen(tt.input)
		if result != tt.expected {
			t.Errorf("floorPowerOfTen(%v) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}
