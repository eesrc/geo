package geometry

import (
	"errors"
	"math"
)

// LineIntersects checks if a line interesects another line
func LineIntersects(line1, line2 []Point) (bool, error) {
	if len(line1) != 2 || len(line2) != 2 {
		return false, errors.New("Line input must consist of array of exactly two points")
	}

	// Check BoundingBox
	line1boundingBox := CalculateBoundingBox(line1)
	line2boundingBox := CalculateBoundingBox(line2)

	if !line1boundingBox.BoundingBoxIntersects(&line2boundingBox) {
		return false, nil
	}

	// Find the four orientations needed for general and special cases
	o1 := orientation(line1[0], line1[1], line2[0])
	o2 := orientation(line1[0], line1[1], line2[1])
	o3 := orientation(line2[0], line2[1], line1[0])
	o4 := orientation(line2[0], line2[1], line1[1])

	// General case
	if o1 != o2 && o3 != o4 {
		return true, nil
	}

	// Special Cases
	// line1[0], line1[1] and line2[0] are colinear and line2[0] lies on segment line1[0]line1[1]
	if o1 == 0 && onSegment(line1[0], line2[0], line1[1]) {
		return true, nil
	}

	// line1[0], line1[1] and line2[1] are colinear and line2[1] lies on segment line1[0]line1[1]
	if o2 == 0 && onSegment(line1[0], line2[1], line1[1]) {
		return true, nil
	}

	// line2[0], line2[1] and line1[0] are colinear and line1[0] lies on segment line2[0]line2[1]
	if o3 == 0 && onSegment(line2[0], line1[0], line2[1]) {
		return true, nil
	}

	// line2[0], line2[1] and line1[1] are colinear and line1[1] lies on segment line2[0]line2[1]
	if o4 == 0 && onSegment(line2[0], line1[1], line2[1]) {
		return true, nil
	}

	return false, nil
}

// Orientation finds the orientation of ordered triplet (p1, p2, p3).
// The function returns following values
// 0 --> p1, p2 and p3 are colinear
// 1 --> Clockwise
// 2 --> Counterclockwise
func orientation(p1, p2, p3 Point) int {
	// See https://www.geeksforgeeks.org/orientation-3-ordered-points/
	// for details of below formula.
	val := Determinate(p1, p2, p3)

	if val == 0 {
		return 0 // colinear
	}

	// clock or counterclock wise
	if val > 0 {
		return 1
	}

	return 2
}

// OnSegment return whether p2 is part of the line p1p3 given given three colinear points
func onSegment(p1, p2, p3 Point) bool {
	if p2.X <= math.Max(p1.X, p3.X) && p2.X >= math.Min(p1.X, p3.X) &&
		p2.Y <= math.Max(p1.Y, p3.Y) && p2.Y >= math.Min(p1.Y, p3.Y) {
		return true
	}

	return false
}
