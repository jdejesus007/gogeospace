package haversine

import (
	"math"

	"github.com/jdejesus007/gogeospace/constants"
	"github.com/jdejesus007/gogeospace/point"
	"github.com/jdejesus007/gogeospace/utils"
)

const (
	// EARTH_RADIUS_CONSTANT Earth Radius used with the Harvesine formula and approximates using a spherical (non-ellipsoid) Earth.
	EARTH_RADIUS_CONSTANT = 6371008.8
)

// lat1, ln2 in degrees
// radius in radians -> ditance / Earth Radius gives radians
func CreateHaversineDisc(lat1, lng1, radius float64) []*point.Point {

	steps := constants.NUM_STEPS_PRECISION               // precision
	radiusRad := radius / float64(EARTH_RADIUS_CONSTANT) // meters
	lat1Rad := utils.DegreesToRadians(lat1)
	lng1Rad := utils.DegreesToRadians(lng1)

	var coordinates []*point.Point
	for i := 0.0; i < steps; i++ {
		bearingRad := utils.DegreesToRadians(float64(i * -360.0 / steps))

		lat2Rad := math.Asin(math.Sin(lat1Rad)*math.Cos(radiusRad) + math.Cos(lat1Rad)*math.Sin(radiusRad)*math.Cos(bearingRad))
		lng2Rad := lng1Rad + math.Atan2(math.Sin(bearingRad)*math.Sin(radiusRad)*math.Cos(lat1Rad), math.Cos(radiusRad)-math.Sin(lat1Rad)*math.Sin(lat2Rad))

		lat2 := utils.RadToDegrees(lat2Rad)
		lng2 := utils.RadToDegrees(lng2Rad)

		coordinates = append(coordinates, &point.Point{Lat: lat2, Lng: lng2})
	}
	return coordinates
}
