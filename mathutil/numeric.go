package mathutil

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Clamp[T Number](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func Abs[T Number](value T) T {
	if value < 0 {
		return -value
	}
	return value
}

// FloorPowerOfTen returns the largest power of ten that is less than or equal to n.
// Not using logarithmic operation bcz its expensive
func FloorPowerOfTen(n int) int {
	if n <= 0 {
		return 1 // return 1 for non-positive numbers
	}
	if n < 10 {
		return 1
	}

	smallTable := []int{1, 10, 100, 1000, 10000, 100000, 1000000}
	for _, v := range smallTable {
		if n < v*10 {
			return v
		}
	}

	// For medium to large values, use the division technique
	value := uint64(n)
	result := 1

	for value >= 10 {
		value /= 10
		result *= 10
	}

	return result
}
