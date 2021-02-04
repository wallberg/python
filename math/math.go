package math

// This package supplements the built-in math package

// MinInt returns the minimum value of the provided ints
func MinInt(is ...int) int {
	min := is[0]
	for _, i := range is[1:] {
		if i < min {
			min = i
		}
	}
	return min
}

// MaxInt returns the maximum value of the provided ints
func MaxInt(is ...int) int {
	max := is[0]
	for _, i := range is[1:] {
		if i > max {
			max = i
		}
	}
	return max
}

// Monus returns max(x-y, 0)
func MonusInt(x int, y int) int {
	return MaxInt(x-y, 0)
}
