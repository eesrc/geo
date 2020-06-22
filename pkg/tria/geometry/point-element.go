package geometry

// PointElement ...
type PointElement struct {
	Prev, Next *PointElement
	Point      Point
}

// Remove ...
func (e *PointElement) Remove() {
	e.Next.Prev, e.Prev.Next = e.Prev, e.Next
}

// CleanUp ...
func (e *PointElement) CleanUp() {
	// Cleanup
	e.Prev = nil
	e.Next = nil
}
