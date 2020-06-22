package geometry

// Polygon contains static and processed information about a polygon
type Polygon struct {
	ID          int64
	Name        string
	Description string
	AlphaShape  []Point
	Triangles   [][]Point
	BoundingBox BoundingBox
	Properties  ShapeProperties
}

// Type returns the shape type
func (polygon *Polygon) Type() string {
	return "polygon"
}

// GetName returns the name of the polygon
func (polygon *Polygon) GetName() string {
	return polygon.Name
}

// GetID Returns the ID of the polygon
func (polygon *Polygon) GetID() int64 {
	return polygon.ID
}

// SetID Sets the ID of the polygon
func (polygon *Polygon) SetID(id int64) {
	polygon.ID = id
}

// GetProperties fetches the properties of the polygon shape
func (polygon *Polygon) GetProperties() ShapeProperties {
	return polygon.Properties
}

// SetProperties sets the properties of the polygon shape
func (polygon *Polygon) SetProperties(shapeProperties ShapeProperties) {
	polygon.Properties = shapeProperties
}

// BoundingBox returns the polygons BoundingBox
func (polygon *Polygon) GetBoundingBox() BoundingBox {
	return polygon.BoundingBox
}

// NewPolygonFromPoints Returns a new Polygon with BoundingBox and AlphaShape, without triangulating points
func NewPolygonFromPoints(points []Point) Polygon {
	return Polygon{
		AlphaShape:  points,
		BoundingBox: CalculateBoundingBox(points),
	}
}

// PointInside checks if a point is inside the given polygon
func (polygon *Polygon) PointInside(point *Point) bool {
	return PointInsidePolygon(point, polygon)
}

// PointInsideBoundingBox checks if a point is inside the given polygon
func (polygon *Polygon) PointInsideBoundingBox(point *Point) bool {
	return polygon.BoundingBox.ContainsPoint(point)
}

// ShapeInside checks if another shape is inside the given polygon
func (polygon *Polygon) ShapeInside(shape Shape) bool {
	switch shape := shape.(type) {
	case *Polygon:
		return polygon.PolygonInside(shape)
	case *Circle:
		return polygon.CircleInside(shape)
	default:
		return false
	}
}

// PolygonInside checks whether a polygon is inside the polygon
func (polygon *Polygon) PolygonInside(polygonCandidate *Polygon) bool {
	for _, point := range polygon.AlphaShape {
		if !polygonCandidate.PointInsideBoundingBox(&point) || !polygonCandidate.PointInside(&point) {
			return false
		}
	}
	return true
}

// CircleInside checks whether a circle is inside the polygon
func (polygon *Polygon) CircleInside(circle *Circle) bool {
	// First check if origo is within BoundingBox. If no, the circle can't be inside polygon
	if polygon.PointInsideBoundingBox(&circle.Origo) {
		// Then we check the outer bounds of and see if the radius is less than the distance between min and max points
		// If yes, we need to check all the points in the polygon alpha shape
		if circle.Origo.DistanceTo(&Point{X: polygon.BoundingBox.MinX, Y: polygon.BoundingBox.MinY}) > circle.Radius &&
			circle.Origo.DistanceTo(&Point{X: polygon.BoundingBox.MaxX, Y: polygon.BoundingBox.MaxY}) > circle.Radius {
			// For each vertice, check if intersecting (ray crossing) and check that points is outside circle radius
			numOfIntersections := 0

			// Creating ray, adding +1 to Y value to ensure intersection of edges
			circleRayCrossing := []Point{circle.Origo, Point{X: polygon.BoundingBox.MaxX + 1, Y: circle.Origo.Y}}

			for i := 0; i < len(polygon.AlphaShape); i++ {
				p1 := polygon.AlphaShape[i]
				p2 := polygon.AlphaShape[(i+1)%len(polygon.AlphaShape)]

				if circle.Origo.DistanceToLine(&p1, &p2) < circle.Radius {
					return false
				}

				intersects, err := LineIntersects(circleRayCrossing, []Point{p1, p2})

				if err != nil {
					return false
				}

				if intersects {
					numOfIntersections++
				}
			}

			// According to ray crossing theorem we know that the circle is inside/outside depending if the number of intersections is odd or even
			return numOfIntersections%2 != 0
		}
	}
	return false
}
