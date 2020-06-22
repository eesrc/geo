package geometry

// PointInsidePolygon checks if given point is present inside polygon by using precalculated triangles
func PointInsidePolygon(p *Point, polygon *Polygon) bool {
	for _, triangle := range polygon.Triangles {
		if PointInsideTriangle(p, &triangle[0], &triangle[1], &triangle[2]) {
			return true
		}
	}
	return false
}

// PointInsideTriangle take a point and an triangle as ta tb tc Point2d
func PointInsideTriangle(p, ta, tb, tc *Point) bool {
	return (tc.X-p.X)*(ta.Y-p.Y)-(ta.X-p.X)*(tc.Y-p.Y) >= 0 &&
		(ta.X-p.X)*(tb.Y-p.Y)-(tb.X-p.X)*(ta.Y-p.Y) >= 0 &&
		(tb.X-p.X)*(tc.Y-p.Y)-(tc.X-p.X)*(tb.Y-p.Y) >= 0
}
