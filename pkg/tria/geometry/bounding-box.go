package geometry

import "math"

// BoundingBox contains information about the Shape bounding box
type BoundingBox struct {
	MaxX float64
	MaxY float64
	MinX float64
	MinY float64
}

// ContainsPoint checks if a Point is within the bounding box
func (boundingBox *BoundingBox) ContainsPoint(p *Point) bool {
	return boundingBox.MinX <= p.X && boundingBox.MaxX >= p.X && boundingBox.MinY <= p.Y && boundingBox.MaxY >= p.Y
}

// BoundingBoxIntersects checks if another BoundingBox intersects with the bounding box
func (boundingBox *BoundingBox) BoundingBoxIntersects(intersectingBoundingBox *BoundingBox) bool {
	return boundingBox.MinX <= intersectingBoundingBox.MaxX && boundingBox.MaxX >= intersectingBoundingBox.MinX &&
		boundingBox.MaxY >= intersectingBoundingBox.MinY && boundingBox.MinY <= intersectingBoundingBox.MaxY
}

// CalculateBoundingBox calculates the bounding box based on given list of Point
func CalculateBoundingBox(points []Point) BoundingBox {
	minX := math.Inf(0)
	maxX := math.Inf(-1)
	minY := math.Inf(0)
	maxY := math.Inf(-1)

	for _, point := range points {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	return BoundingBox{
		MinX: minX,
		MaxX: maxX,
		MinY: minY,
		MaxY: maxY,
	}
}
