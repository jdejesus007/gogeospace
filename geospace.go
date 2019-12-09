package gogeospace

import (
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/jdejesus007/gogeos/geos"
	"github.com/jdejesus007/gogeospace/haversine"
	"github.com/jdejesus007/gogeospace/point"
	"github.com/jdejesus007/gogeospace/vincenty"
	"github.com/pkg/errors"
)

// GetIntersectedPolygonByPolygonAndCenterPointRadiusHaveriseDisc returns one polygon consisting of
// an individual intersected polygon and a disc derived of the passed in center
// point and radius with haversine algorithm
// Params:
// Coordinates forming a polygon slice of lat,lng in degress
// Lat center point lat in degrees
// Lng center point lng in degrees
// Radius off center point to create spherical disc or circle in meters
func GetIntersectedPolygonByPolygonAndCenterPointRadiusHaveriseDisc(
	polyCoords []*point.Point,
	lat float32,
	lng float32,
	radius float64) (coordinates []*point.Point, err error) {

	// Catch internal C library panics
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = errors.Wrap(fmt.Errorf("Error: %v", e), fmt.Sprintf("Debug Stack: %s", string(debug.Stack())))
				return
			}
			err = errors.Wrap(err, fmt.Sprintf("Debug Stack: %s", string(debug.Stack())))
		}
	}()

	dotPolygon, err := getGeosPolygonFromCoordinates(polyCoords)
	if err != nil {
		return nil, err
	}

	// Convert to spherical radius -> radians = distance / earth radius
	// C lib has problems with gaps around polygon edges
	polyCoordinates := haversine.CreateHaversineDisc(float64(lat), float64(lng), radius)

	intersectedPolyCoords, err := processPolyCoordinates(polyCoordinates, dotPolygon)
	if err != nil {
		return nil, err
	}

	for _, coords := range intersectedPolyCoords {
		intersectedPolyPoints := strings.Split(coords, ",")
		log.Println("Points: ", intersectedPolyPoints)
		for _, p := range intersectedPolyPoints {
			latLng := strings.Split(strings.TrimSpace(p), " ")
			lat, _ := strconv.ParseFloat(strings.Trim(latLng[0], " "), 10)
			lng, _ := strconv.ParseFloat(strings.Trim(latLng[1], " "), 10)
			coordinates = append(coordinates, &point.Point{Lat: lat, Lng: lng})

		}
	}

	return coordinates, nil
}

// GetIntersectedPolygonByPolygonAndCenterPointRadiusVincentyDisc returns one polygon consisting of
// an individual intersected polygon and a disc derived of the passed in center
// point and radius with vincenty algorithm
// Params:
// Coordinates forming a polygon slice of lat,lng in degress
// Lat center point lat in degrees
// Lng center point lng in degrees
// Radius off center point to create spherical disc or circle in meters
func GetIntersectedPolygonByPolygonAndCenterPointRadiusVincentyDisc(
	polyCoords []*point.Point,
	lat float32,
	lng float32,
	radius float64) (coordinates []*point.Point, err error) {

	// Catch internal C library panics
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = errors.Wrap(fmt.Errorf("Error: %v", e), fmt.Sprintf("Debug Stack: %s", string(debug.Stack())))
				return
			}
			err = errors.Wrap(err, fmt.Sprintf("Debug Stack: %s", string(debug.Stack())))
		}
	}()

	dotPolygon, err := getGeosPolygonFromCoordinates(polyCoords)
	if err != nil {
		return nil, err
	}

	// Convert to spherical radius -> radians = distance / earth radius
	// C lib has problems with gaps around polygon edges
	polyCoordinates := vincenty.CreateVincentyDisc(float64(lat), float64(lng), radius) // accurate to within 0.5 mm distance or 0.000015″ of bearing

	intersectedPolyCoords, err := processPolyCoordinates(polyCoordinates, dotPolygon)
	if err != nil {
		return nil, err
	}

	for _, coords := range intersectedPolyCoords {
		intersectedPolyPoints := strings.Split(coords, ",")
		log.Println("Points: ", intersectedPolyPoints)
		for _, p := range intersectedPolyPoints {
			latLng := strings.Split(strings.TrimSpace(p), " ")
			lat, _ := strconv.ParseFloat(strings.Trim(latLng[0], " "), 10)
			lng, _ := strconv.ParseFloat(strings.Trim(latLng[1], " "), 10)
			coordinates = append(coordinates, &point.Point{Lat: lat, Lng: lng})

		}
	}

	return coordinates, nil
}

func processPolyCoordinates(polyCoordinates []*point.Point, dotPolygon *geos.Geometry) ([]string, error) {
	// Generic collection of points - convert to coordinate string
	var pointsStr string
	for _, p := range polyCoordinates {
		pointsStr += fmt.Sprintf("%f %f, ", p.Lat, p.Lng)
	}

	pointsStr = pointsStr[:len(pointsStr)-2]
	sepPoints := strings.Split(pointsStr, ",")

	// NOTE:
	// Split by comma - repeat the first point to close polygon
	// If we do not do this, it will panic with: geos: IllegalArgumentException: Points of LinearRing do not form a closed linestring
	// Per Geos C++ Port of Original JTP - Java Topology Suite - a valid polygon
	// is a closed circuit with exact points at the beginning and end of the
	// polygon points sequence
	rawCirclePolygonStr := fmt.Sprintf("POLYGON ((%s))", pointsStr+", "+sepPoints[0]) // take the first and attach the end to close the polygon

	// Final intersected polygon - do this for DOT with service radius only
	circlePoly := geos.Must(geos.FromWKT(rawCirclePolygonStr))
	intersectedPoly := geos.Must(dotPolygon.Intersection(circlePoly))

	polyType, err := intersectedPoly.Type()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting polygon type")
	}

	// If nonintersecting - return empty to skip area
	if intersectedPoly.String()[len(intersectedPoly.String())-5:] == "EMPTY" {
		log.Println("empty")
		return nil, nil
	}

	// Extract and build up coordinates
	var intersectedPolyCoords []string
	switch polyType {
	case geos.POLYGON:
		polyStr := intersectedPoly.String()
		polyStr = strings.Replace(polyStr, "), (", ", ", -1) // sanitize this format ), (
		intersectedPolyCoords = append(intersectedPolyCoords, polyStr[10:len(polyStr)-2])
	case geos.MULTIPOLYGON:
		// We have multi polygon when we have lines crossing - due to gaps initially
		n, err := intersectedPoly.NGeometry()
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < n; i++ {
			geo := geos.Must(intersectedPoly.Geometry(i))
			points := geo.String()[10 : len(geo.String())-2]
			points = strings.Replace(points, "), (", ", ", -1) // sanitize this format ), (
			intersectedPolyCoords = append(intersectedPolyCoords, points)
		}
	default:
		log.Fatalln("Unknown type", polyType, intersectedPoly)
	}

	return intersectedPolyCoords, nil
}

// Expected format - slice of coordinate points
func getGeosPolygonFromCoordinates(coordinates []*point.Point) (geosPoly *geos.Geometry, err error) {
	var points string
	for _, point := range coordinates {
		points += fmt.Sprintf("%f %f, ", point.Lat, point.Lng)
	}
	points = points[:len(points)-2]

	// NOTE:
	// Split by comma - repeat the first point to close polygon
	// If we do not do this, it will panic with: geos: IllegalArgumentException: Points of LinearRing do not form a closed linestring
	// Per Geos C++ Port of Original JTP - Java Topology Suite - a valid polygon
	// is a closed circuit with exact points at the beginning and end of the
	// polygon points sequence
	sepPoints := strings.Split(points, ",")
	outputt := fmt.Sprintf("POLYGON ((%s))", points+", "+sepPoints[0])

	// This will panic if C library returns non-nil error - catch in calling
	// method
	geosPoly = geos.Must(geos.FromWKT(outputt))

	return geosPoly, nil
}
