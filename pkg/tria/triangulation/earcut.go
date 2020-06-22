package triangulation

import (
	"errors"

	"github.com/eesrc/geo/pkg/tria/geometry"
)

// Error contains information about the state when a triangulation happened
type Error struct {
	// Description short description of error
	Description string
	// Ears ears of the triangulation when it failed
	Ears [][]geometry.Point
	// PolygonWhenError The failing polygon upon failure
	PolygonWhenError []geometry.Point
}

// newTriangulationError creates a Error specific to a failed triangulation which contains ears and polygon when the error occured
func newTriangulationError(text string, ears [][]geometry.Point, poly []geometry.Point) error {
	return Error{
		Description:      text,
		Ears:             ears,
		PolygonWhenError: poly,
	}
}

func (e Error) Error() string {
	return e.Description
}

// isEar checks if an ear candidate is an ear
func isEar(earCandidate *geometry.PointElement) bool {
	a, b, c := earCandidate.Prev.Point, earCandidate.Point, earCandidate.Next.Point

	// The triangle testing for needs to be convex to be clippable
	if !geometry.IsConvex(a, b, c) {
		return false
	}

	// Start with the next point after ear candidate
	verticePoint := earCandidate.Next.Next

	// Search through all points to see if they exist within triangle A, B, C. If yes, it's not an ear.
	// Limit iteration on previous element as we've looped through all elements if satisfied.
	for verticePoint != earCandidate.Prev {
		if geometry.PointInsideTriangle(&verticePoint.Point, &a, &b, &c) {
			return false
		}
		verticePoint = verticePoint.Next
	}

	return true
}

// GetEars Returns polygon ears by using the ear cutting algorithm (n^2)
func GetEars(polygon []geometry.Point) ([][]geometry.Point, error) {
	polygonSize := len(polygon)

	if polygonSize < 3 {
		return [][]geometry.Point{}, errors.New("Polygon is smaller than 3 points")
	}

	if polygonSize == 3 {
		return [][]geometry.Point{polygon}, nil
	}

	// A triangulated polygon consists of n-2 triangles where n is the number of vertices
	triangles := make([][]geometry.Point, polygonSize-2)

	// Allocation and manual initialization of point elements
	pointElements := make([]geometry.PointElement, polygonSize)
	pointElements[0].Prev, pointElements[0].Next = &pointElements[polygonSize-1], &pointElements[1]
	pointElements[0].Point = polygon[0]

	for i := 1; i < polygonSize-1; i++ {
		pointElements[i].Prev, pointElements[i].Next = &pointElements[i-1], &pointElements[i+1]
		pointElements[i].Point = polygon[i]
	}

	pointElements[polygonSize-1].Prev, pointElements[polygonSize-1].Next = &pointElements[polygonSize-2], &pointElements[0]
	pointElements[polygonSize-1].Point = polygon[polygonSize-1]

	// Start triangulation
	ear := &pointElements[0]
	stop := ear.Prev

	triangleNumber := 0

	// While we have more than three points
	for ear.Prev != ear.Next {
		// The triangulation has looped since the last found ear, meaning it has failed.
		// Return a triangulationError for debugging
		if ear == stop {
			var pointsWhenError []geometry.Point

			// Unwrap error points.
			// Start with the next point after ear candidate
			r := ear.Next.Next
			for r != ear.Prev {
				pointsWhenError = append(pointsWhenError, r.Point)
				r = r.Next
			}

			return [][]geometry.Point{}, newTriangulationError("Looped polygon vertices. Couldn't triangulate the whole polygon", triangles[:triangleNumber], pointsWhenError)
		}

		if isEar(ear) {
			// Check if the ear has an area
			if geometry.PolygonArea([]geometry.Point{ear.Prev.Point, ear.Point, ear.Next.Point}) > 0 {
				// We're dealing with an ear, add triangle
				triangles[triangleNumber] = []geometry.Point{ear.Prev.Point, ear.Point, ear.Next.Point}
				triangleNumber++
			}

			// Remove ear edge from linked list
			ear.Remove()

			// Set new ear and stop ear
			ear = ear.Next
			stop = ear.Prev
		}

		ear = ear.Next
	}

	return triangles[:triangleNumber], nil
}

// TriangulateByEarCut Triangulates a list of points returning a list of triangles of Point2d
func TriangulateByEarCut(polygon []geometry.Point) ([][]geometry.Point, error) {
	normalizedPoints := geometry.NormalizePoints(polygon)

	return GetEars(normalizedPoints)
}
