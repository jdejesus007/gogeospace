package vincenty

import (
	"math"

	"github.com/jdejesus007/gogeospace/constants"
	"github.com/jdejesus007/gogeospace/point"
	"github.com/jdejesus007/gogeospace/utils"
)

const (
	// WGS-84 ellipsoid params
	a        = 6378137        // Semi-Major Axis in meters
	b        = 6356752.314245 // Semi-Minor Axis in meters
	f        = (a - b) / a    // Flattening
	aSquared = a * a
	bSquared = b * b
)

// CreateVincentyDisc creates a disc with center lat1, lng1, and radius in meters
func CreateDisc(lat1, lng1, radius float64) []*point.Point {
	// all going in as degrees and meters
	steps := constants.NUM_STEPS_PRECISION // precision
	var coordinates []*point.Point
	for i := 0.0; i < steps; i++ {
		startBearing := float64(i * -360.0 / steps)
		lat2, lng2, _ := CalculateVincentyCoordinate(lat1, lng1, radius, startBearing)
		coordinates = append(coordinates, &point.Point{Lat: lat2, Lng: lng2})
	}
	return coordinates
}

// CalculateVincentyCoordinate gets a point on the disc given center in
// degrees, radius distance in meters, and bearing in degrees
// Returns latitude2, longitude2, and ending bearing in degrees
func CalculateVincentyCoordinate(lat1, lng1, radius, startBearing float64) (float64, float64, float64) {
	phi1 := utils.DegreesToRadians(lat1)
	alpha1 := utils.DegreesToRadians(startBearing)
	cosAlpha1 := math.Cos(alpha1)
	sinAlpha1 := math.Sin(alpha1)
	s := radius
	tanU1 := (1.0 - f) * math.Tan(phi1)
	cosU1 := 1.0 / math.Sqrt(1.0+tanU1*tanU1)
	sinU1 := tanU1 * cosU1

	// eq. 1
	sigma1 := math.Atan2(tanU1, cosAlpha1)

	// eq. 2
	sinAlpha := cosU1 * sinAlpha1

	sin2Alpha := sinAlpha * sinAlpha
	cos2Alpha := 1 - sin2Alpha
	uSquared := cos2Alpha * (aSquared - bSquared) / bSquared

	// eq. 3
	A := 1 + (uSquared/16384)*(4096+uSquared*(-768+uSquared*(320-175*uSquared)))

	// eq. 4
	B := (uSquared / 1024) * (256 + uSquared*(-128+uSquared*(74-47*uSquared)))

	// iterate until there is a negligible change in sigma
	var (
		deltaSigma  float64
		sOverbA     = s / (b * A)
		sigma       = sOverbA
		sinSigma    float64
		prevSigma   = sOverbA
		sigmaM2     float64
		cosSigmaM2  float64
		cos2SigmaM2 float64
	)

	for {
		// eq. 5
		sigmaM2 = 2.0*sigma1 + sigma
		cosSigmaM2 = math.Cos(sigmaM2)
		cos2SigmaM2 = cosSigmaM2 * cosSigmaM2
		sinSigma = math.Sin(sigma)
		cosSigma := math.Cos(sigma)

		// eq. 6
		deltaSigma = B * sinSigma * (cosSigmaM2 + (B/4.0)*(cosSigma*(-1+2*cos2SigmaM2)-(B/6.0)*cosSigmaM2*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM2)))

		// eq. 7
		sigma = sOverbA + deltaSigma

		// break after converging to tolerance
		if math.Abs(sigma-prevSigma) < 0.0000000000001 {
			break
		}

		prevSigma = sigma
	}

	sigmaM2 = 2.0*sigma1 + sigma
	cosSigmaM2 = math.Cos(sigmaM2)
	cos2SigmaM2 = cosSigmaM2 * cosSigmaM2

	cosSigma := math.Cos(sigma)
	sinSigma = math.Sin(sigma)

	// eq. 8
	phi2 := math.Atan2(sinU1*cosSigma+cosU1*sinSigma*cosAlpha1, (1.0-f)*math.Sqrt(sin2Alpha+math.Pow(sinU1*sinSigma-cosU1*cosSigma*cosAlpha1, 2.0)))

	// eq. 9 pole crossing defect fixed
	lambda := math.Atan2(sinSigma*sinAlpha1, (cosU1*cosSigma - sinU1*sinSigma*cosAlpha1))

	// eq. 10
	C := (f / 16) * cos2Alpha * (4 + f*(4-3*cos2Alpha))

	// eq. 11
	L := lambda - (1-C)*f*sinAlpha*(sigma+C*sinSigma*(cosSigmaM2+C*cosSigma*(-1+2*cos2SigmaM2)))

	// eq. 12
	alpha2 := math.Atan2(sinAlpha, -sinU1*sinSigma+cosU1*cosSigma*cosAlpha1)

	// end coordinate bearing result
	endBearing := utils.RadToDegrees(alpha2)

	// coordinate result
	latitude := utils.RadToDegrees(phi2)
	longitude := lng1 + utils.RadToDegrees(L)

	return latitude, longitude, endBearing
}
