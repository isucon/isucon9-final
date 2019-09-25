package util

import "math/rand"

func RandRangeIntn(min, max int) int {
	return rand.Intn(max-min) + min
}
