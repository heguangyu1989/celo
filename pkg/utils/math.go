package utils

import "time"

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SecondsToDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
