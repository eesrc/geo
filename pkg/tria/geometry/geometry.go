package geometry

import (
	"math"
)

// Determinate calculates the determinate of a triangle
func Determinate(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-b.Y) - (c.X-b.X)*(b.Y-a.Y)
}

// IsConvex checks if a triangle is convex (angle less than 180deg) based on its determinate
func IsConvex(a, b, c Point) bool {
	return Determinate(a, b, c) > 0
}

// IsClockwise checks if a polygon is clockwise or not
func IsClockwise(poly []Point) bool {
	if len(poly) < 3 {
		return false
	}
	// initialize sum with first to last element
	sum := (poly[0].X - poly[len(poly)-1].X) * (poly[0].Y + poly[len(poly)-1].Y)
	// iterate over all other points (0 to n-1)
	for i := 0; i < len(poly)-1; i++ {
		sum += (poly[i+1].X - poly[i].X) * (poly[i+1].Y + poly[i].Y)
	}

	return sum > 0
}

// PolygonArea calculates an area of a polygon
func PolygonArea(data []Point) float64 {
	area := 0.0
	for i, j := 0, len(data)-1; i < len(data); i++ {
		area += data[i].X*data[j].Y - data[i].Y*data[j].X
		j = i
	}
	return math.Abs(area / 2)
}
