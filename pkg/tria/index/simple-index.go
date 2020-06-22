package index

import (
	"errors"
	"sync"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/tria/geometry"
)

// SimpleIndex is a simple store with no optmizations and implements Store from Tria
type SimpleIndex struct {
	Shapes []geometry.Shape

	storeMutex *sync.Mutex
}

// NewSimpleIndex returns an initialized SimpleIndex
func NewSimpleIndex() *SimpleIndex {
	store := SimpleIndex{}
	store.storeMutex = &sync.Mutex{}

	return &store
}

// NewSimpleIndexFromModel returns an initialized SimpleIndex with index populated from a list of shapes
func NewSimpleIndexFromModel(shapeModels []model.Shape) *SimpleIndex {
	index := NewSimpleIndex()

	shapes := make([]geometry.Shape, len(shapeModels))

	for i, shapeModel := range shapeModels {
		shapes[i] = shapeModel.Shape
	}

	index.AddShapes(shapes)

	return index
}

// AddShape ...
func (store *SimpleIndex) AddShape(shape geometry.Shape) {
	store.storeMutex.Lock()
	defer store.storeMutex.Unlock()
	store.Shapes = append(store.Shapes, shape)
}

// AddShapes ...
func (store *SimpleIndex) AddShapes(shapes []geometry.Shape) {
	store.storeMutex.Lock()
	defer store.storeMutex.Unlock()
	store.Shapes = append(store.Shapes, shapes...)
}

// GetShapeByName ...
func (store *SimpleIndex) GetShapeByName(shapeName string) (geometry.Shape, error) {
	for _, shape := range store.Shapes {
		if shape.GetName() == shapeName {
			return shape, nil
		}
	}

	return &geometry.Circle{}, errors.New("No shape found with name " + shapeName)
}

// RemoveShapeByName ...
func (store *SimpleIndex) RemoveShapeByName(shapeName string) (geometry.Shape, error) {
	for i, shape := range store.Shapes {
		if shape.GetName() == shapeName {
			store.storeMutex.Lock()
			defer store.storeMutex.Unlock()

			shapeToRemove := shape

			store.Shapes = append(store.Shapes[:i], store.Shapes[i+1:]...)

			return shapeToRemove, nil
		}
	}

	return &geometry.Circle{}, errors.New("No shape found with name " + shapeName)
}

// FindShapesWhichContainsPoint ...
func (store *SimpleIndex) FindShapesWhichContainsPoint(point geometry.Point) []geometry.Shape {
	var matchingShapes []geometry.Shape

	for _, shape := range store.Shapes {
		if shape.PointInsideBoundingBox(&point) && shape.PointInside(&point) {
			matchingShapes = append(matchingShapes, shape)
		}
	}

	return matchingShapes
}

// FindShapesWhichContainsShape ...
func (store *SimpleIndex) FindShapesWhichContainsShape(shape geometry.Shape) []geometry.Shape {
	var matchingShapes []geometry.Shape

	for _, shape := range store.Shapes {
		if shape.ShapeInside(shape) {
			matchingShapes = append(matchingShapes, shape)
		}
	}

	return matchingShapes

}
