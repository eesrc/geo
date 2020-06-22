package geometry

import "testing"

// Insert ...
func insert(p Point, e *PointElement) *PointElement {
	newEl := PointElement{Point: p}

	if e != nil {
		newEl.Next = e.Next
		newEl.Prev = e
		e.Next.Prev = &newEl
		e.Next = &newEl
	} else {
		newEl.Prev = &newEl
		newEl.Next = &newEl
	}
	return &newEl
}
func TestElement_Remove(t *testing.T) {
	t.Run("removal", func(t *testing.T) {
		e1 := insert(Point{0, 0}, nil)
		e2 := insert(Point{1, 0}, e1)
		e3 := insert(Point{1, 1}, e2)

		if e1.Next != e2 || e2.Next != e3 {
			t.Error("Wrong insert")
		}

		e2.Remove()

		if e1.Next != e3 || e3.Prev != e1 {
			t.Error("Removal did not connect outer nodes")
		}

		if e2.Prev != e1 || e2.Next != e3 {
			t.Error("Removal did not preserve connections")
		}
	})

	t.Run("remove edge", func(t *testing.T) {
		e1 := insert(Point{1, 1}, nil)
		e2 := insert(Point{2, 2}, e1)
		e3 := insert(Point{3, 3}, e2)

		e3.Remove()

		if e2.Next != e1 || e1.Prev != e2 {
			t.Error("Edge removal makes incorrect connection")
		}
	})
}
