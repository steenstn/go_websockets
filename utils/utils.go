package utils

func Clamp(input int, min int, max int) int {
	if input < min {
		return min
	} else if input > max {
		return max
	} else {
		return input
	}
}
