package utils

func Clamp(val, low, high int64) int64 {
	return max(low, min(val, high))
}
