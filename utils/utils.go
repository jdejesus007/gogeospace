package utils

import (
	"math"
)

func RadToDegrees(r float64) float64 {
	degrees := math.Mod(r, (2.0 * math.Pi))
	return degrees * 180.0 / math.Pi
}

func DegreesToRadians(d float64) float64 {
	radians := math.Mod(d, 360.0)
	return radians * math.Pi / 180.0
}

func LengthToRadians(radius, earthRadius float64) float64 {
	return (radius / earthRadius) // both in meters
}
