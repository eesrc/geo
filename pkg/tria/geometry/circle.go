package geometry

// Circle contains information about circle
type Circle struct {
	ID         int64
	Name       string
	Origo      Point
	Radius     float64
	Properties ShapeProperties
}

// GetName returns the circle name
func (circle *Circle) GetName() string {
	return circle.Name
}

// Type returns the shape type
func (circle *Circle) Type() string {
	return "circle"
}

// GetID Returns the ID of the circle
func (circle *Circle) GetID() int64 {
	return circle.ID
}

// SetID Sets the ID of the circle
func (circle *Circle) SetID(id int64) {
	circle.ID = id
}

// GetProperties fetches the properties of the circle shape
func (circle *Circle) GetProperties() ShapeProperties {
	return circle.Properties
}

// SetProperties sets the properties of the circle shape
func (circle *Circle) SetProperties(shapeProperties ShapeProperties) {
	circle.Properties = shapeProperties
}

// BoundingBox returns the circles BoundingBox
func (circle *Circle) GetBoundingBox() BoundingBox {
	return BoundingBox{
		MinX: circle.Origo.X - circle.Radius,
		MinY: circle.Origo.Y - circle.Radius,
		MaxX: circle.Origo.X + circle.Radius,
		MaxY: circle.Origo.Y + circle.Radius,
	}
}

// PointInside checks if a point is inside circle
func (circle *Circle) PointInside(point *Point) bool {
	return circle.Origo.DistanceTo(point) <= circle.Radius
}

// PointInsideBoundingBox checks if a point is inside the circle
func (circle *Circle) PointInsideBoundingBox(point *Point) bool {
	return circle.Origo.DistanceTo(point) <= circle.Radius
}

// ShapeInside checks if a shape is inside the circle
func (circle *Circle) ShapeInside(shape Shape) bool {
	switch shape := shape.(type) {
	case *Polygon:
		return circle.PolygonInside(shape)
	case *Circle:
		return circle.CircleInside(shape)
	default:
		return false
	}
}

// PolygonInside checks whether a polygon is inside the circle
func (circle *Circle) PolygonInside(polygon *Polygon) bool {
	if circle.PointInside(&Point{X: polygon.BoundingBox.MinX, Y: polygon.BoundingBox.MinY}) ||
		circle.PointInside(&Point{X: polygon.BoundingBox.MaxX, Y: polygon.BoundingBox.MaxY}) {
		for _, point := range polygon.AlphaShape {
			if !circle.PointInside(&point) {
				return false
			}
		}

		return true
	}
	return false
}

// CircleInside checks whether a circle is inside the circle
// Expects the Radius to be normalized according to X,Y-grid
func (circle *Circle) CircleInside(circleCandidate *Circle) bool {
	distOrigo := circle.Origo.DistanceTo(&circleCandidate.Origo)

	return (distOrigo + circleCandidate.Radius) <= circle.Radius
}
