package geometry

import (
	"errors"
	"math"
)

// Point is a 2d point representation
type Point struct {
	X float64
	Y float64
}

// NewPointFromFloat takes a list of float64 which represents X and Y and returns a single Point
func NewPointFromFloat(floatList []float64) (Point, error) {
	if len(floatList) != 2 {
		return Point{}, errors.New("You need exactly two floats to create a single Point. If you want multiple, use NewPointsFromFloat() instead")
	}

	return Point{X: floatList[0], Y: floatList[1]}, nil
}

// NewPointsFromFloat takes a list of float64 and returns a list of Points
func NewPointsFromFloat(floatList []float64) ([]Point, error) {
	if (len(floatList) % 2) != 0 {
		return []Point{}, errors.New("You need exactly two floats to create a single Point and the list needs to be listLength % 2 == 0")
	}

	points := make([]Point, len(floatList)/2)

	for i := 0; i < len(floatList); i += 2 {
		points[i/2] = Point{X: floatList[i], Y: floatList[i+1]}
	}

	return points, nil
}

// Equal checks if a point is equal to another
func (p1 *Point) Equal(p2 *Point) bool {
	return (p1.X == p2.X && p1.Y == p2.Y)
}

// Cross takes the cross product against another point
func (p1 *Point) Cross(p2 *Point) float64 {
	return p1.X*p2.Y - p1.Y*p2.X
}

// Dot returns the dot product against another point
func (p1 *Point) Dot(p2 *Point) float64 {
	return (p1.X * p2.X) + (p1.Y * p2.Y)
}

// DistanceTo calculates squared distance between two points.
func (p1 *Point) DistanceTo(p2 *Point) float64 {
	return math.Sqrt((p1.X-p2.X)*(p1.X-p2.X) + (p1.Y-p2.Y)*(p1.Y-p2.Y))
}

// PlusPoint Adds points together and returns a new Point instance
func (p1 *Point) PlusPoint(p2 *Point) Point {
	return Point{X: p1.X + p2.X, Y: p1.Y + p2.Y}
}

// MinusPoint Subtracts points together and returns a new Point instance
func (p1 *Point) MinusPoint(p2 *Point) Point {
	return Point{X: p1.X - p2.X, Y: p1.Y - p2.Y}
}

// Times Multiplies a value together with the point and returns a new Point instance
func (p1 *Point) Times(value float64) Point {
	return Point{X: p1.X * value, Y: p1.Y * value}
}

// Divide Divides a value with the point and returns a new Point instance
func (p1 *Point) Divide(value float64) Point {
	return Point{X: p1.X / value, Y: p1.Y / value}
}

// DistanceToLine Checks the point' distance towards a line by creating a 90 deg normal
// between p1 and the line
func (p1 *Point) DistanceToLine(a, b *Point) float64 {
	// A + dot(AP,AB) / dot(AB,AB) * AB
	AP := p1.MinusPoint(a)
	AB := b.MinusPoint(a)

	dotAPAB := AP.Dot(&AB)
	dotABAB := AB.Dot(&AB)

	// Get vector
	ABDotted := AB.Times(dotAPAB / dotABAB)

	// Project vector from A
	normalLineIntersection := a.PlusPoint(&ABDotted)

	// Check whether the normal is actually on the line, otherwise check towards one of the end points
	if normalLineIntersection.DistanceTo(a)+normalLineIntersection.DistanceTo(b) == a.DistanceTo(b) {
		return normalLineIntersection.DistanceTo(p1)
	}

	distanceToA := p1.DistanceTo(a)
	distanceToB := p1.DistanceTo(b)

	return math.Min(distanceToA, distanceToB)
}
