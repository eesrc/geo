package geometry

// NormalizePoints takes in an array of Point2d and removes redundant points from the list
func NormalizePoints(points []Point) []Point {
	if len(points) < 3 {
		return points
	}

	pointsToNormalize := make([]Point, len(points))
	copy(pointsToNormalize, points)

	pointsLength := len(pointsToNormalize)

	// Normalize: Remove duplicate start and ending
	if pointsToNormalize[0].Equal(&pointsToNormalize[pointsLength-1]) {
		pointsToNormalize = pointsToNormalize[:pointsLength-1]
		pointsLength--
	}

	// Normalize: Reverse order if clockwise
	if IsClockwise(pointsToNormalize) {
		for i := 0; i < pointsLength/2; i++ {
			pointsToNormalize[i], pointsToNormalize[pointsLength-i-1] = pointsToNormalize[pointsLength-i-1], pointsToNormalize[i]
		}
	}

	for i := 0; i < pointsLength+1; i += 1 {
		tA := pointsToNormalize[(i)%pointsLength]
		tB := pointsToNormalize[(i+1)%pointsLength]
		tC := pointsToNormalize[(i+2)%pointsLength]

		// Normalize: Remove redundant points on same line
		if tB.DistanceTo(&tA)+tB.DistanceTo(&tC) == tA.DistanceTo(&tC) && PolygonArea([]Point{tA, tB, tC}) == 0 {
			if pointsLength < i+2 {
				pointsToNormalize = append(pointsToNormalize[0:(i+1)%pointsLength], pointsToNormalize[(i+2)%pointsLength:]...)
			} else {
				pointsToNormalize = append(pointsToNormalize[:i+1], pointsToNormalize[(i+2):]...)
			}
			pointsLength--
			i--
		}
	}

	return pointsToNormalize
}
