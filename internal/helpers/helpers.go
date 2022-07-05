package helpers

import "math"

func round(num float32) int {
	return int(num + float32(math.Copysign(0.5, float64(num))))
}

func ToFixed(num float32, precision int) float32 {
	output := float32(math.Pow(10, float64(precision)))
	return float32(round(num*output)) / output
}
