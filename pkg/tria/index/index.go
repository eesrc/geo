package index

import "github.com/eesrc/geo/pkg/tria/geometry"

// TriaIndex is a interface for handling searching and storing of shapes
type TriaIndex interface {
	// AddShape adds a shape to the current TriaIndex
	AddShape(geometry.Shape)
	// AddShapes add multiple shapes to the current TriaIndex
	AddShapes([]geometry.Shape)
	// GetShapeByName returns a shape by name if found, otherwise error
	GetShapeByName(string) (geometry.Shape, error)
	// RemoveShapeByName removes a shape by name if found and returns it, otherwise error
	RemoveShapeByName(string) (geometry.Shape, error)

	// FindShapesWhichContains checks the store shapes if it contains with given point
	FindShapesWhichContainsPoint(geometry.Point) []geometry.Shape
	// FindShapesWhichContainsShape checks the store shapes if it contains with given shape
	FindShapesWhichContainsShape(geometry.Shape) []geometry.Shape
}
