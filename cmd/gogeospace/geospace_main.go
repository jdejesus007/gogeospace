package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	geospace "github.com/jlectronix/gogeospace"
	"github.com/jlectronix/gogeospace/point"
)

func main() {
	strCoord := "26.88,-80.5443|26.73,-80.22|26.423,-80.421"

	pointsArr := strings.Split(strCoord, "|")
	var coordinates []*point.Point
	for _, p := range pointsArr {
		singleLatLng := strings.Split(p, ",")
		lat, _ := strconv.ParseFloat(strings.Trim(singleLatLng[0], " "), 10)
		lng, _ := strconv.ParseFloat(strings.Trim(singleLatLng[1], " "), 10)
		coordinates = append(coordinates, &point.Point{Lat: lat, Lng: lng})
	}

	intersectedPolyPoints, err := geospace.GetIntersectedPolygonByPolygonAndCenterPointRadiusHaveriseDisc(coordinates, 26.43, -80.32, 7000)
	if err != nil {
		log.Println("Failed intersecting polygon with haversine disc")
	}

	if len(intersectedPolyPoints) > 0 {
		var transformedPoints string
		for _, p := range intersectedPolyPoints {
			transformedPoints += fmt.Sprintf("%f,%f|", p.Lat, p.Lng) // transform to expected format 26,80|29,81
		}

		if len(transformedPoints) > 0 {
			transformedPoints = transformedPoints[:len(transformedPoints)-1] // remove last pipe
		}

		log.Println("Final with Harversine: ", transformedPoints)
	}

	intersectedPolyPoints, err = geospace.GetIntersectedPolygonByPolygonAndCenterPointRadiusVincentyDisc(coordinates, 26.43, -80.32, 7000)
	if err != nil {
		log.Println("Failed intersecting polygon with vincenty disc")
	}

	if len(intersectedPolyPoints) > 0 {
		var transformedPoints string
		for _, p := range intersectedPolyPoints {
			transformedPoints += fmt.Sprintf("%f,%f|", p.Lat, p.Lng) // transform to expected format 26,80|29,81
		}

		if len(transformedPoints) > 0 {
			transformedPoints = transformedPoints[:len(transformedPoints)-1] // remove last pipe
		}

		log.Println("Final with Vincenty: ", transformedPoints)
	}
}
