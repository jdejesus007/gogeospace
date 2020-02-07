package polygon

import (
	"fmt"
	"runtime/debug"

	"github.com/jdejesus007/gogeospace/point"
	"github.com/paulsmith/gogeos/geos"
	"github.com/pkg/errors"
)

// error if polygons intersect
func DoPolygonsIntersect(coordinatesA, coordinatesB []*point.Point) (intersects bool, err error) {
	// Catch internal C library panics
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			intersects = false
			err, ok = e.(error)
			if !ok {
				err = errors.Wrap(fmt.Errorf("Error: %v", e), fmt.Sprintf("Debug Stack: %s", string(debug.Stack())))
				return
			}
			err = errors.Wrap(err, fmt.Sprintf("Debug Stack: %s", string(debug.Stack())))
		}
	}()

	dotPolygonA, err := getGeosPolygonFromCoordinates(coordinatesA)
	if err != nil {
		return false, err
	}

	dotPolygonB, err := getGeosPolygonFromCoordinates(coordinatesB)
	if err != nil {
		return false, err
	}

	intersectedPoly := geos.Must(dotPolygonA.Intersection(dotPolygonB))

	// If nonintersecting - return empty to skip area
	if intersectedPoly.String()[len(intersectedPoly.String())-5:] == "EMPTY" {
		return false, nil
	}

	return true, nil
}
