package index

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dhconnelly/rtreego"
	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/tria/geometry"
	log "github.com/sirupsen/logrus"
)

type RTreeIndex struct {
	tree        *rtreego.Rtree
	treeObjects []*rTreeObject

	storeMutex *sync.Mutex
}

type rTreeObject struct {
	shape geometry.Shape
	rect  *rtreego.Rect
}

func (rTreeObj *rTreeObject) Bounds() *rtreego.Rect {
	return rTreeObj.rect
}

func NewRTreeIndex() *RTreeIndex {
	rtree := RTreeIndex{}

	rtree.tree = rtreego.NewTree(2, 25, 50)
	rtree.storeMutex = &sync.Mutex{}

	return &rtree
}

func NewRTreeIndexFromModel(shapeModels []model.Shape) *RTreeIndex {
	index := NewRTreeIndex()

	then := time.Now()

	for _, shapeModel := range shapeModels {
		index.AddShape(shapeModel.Shape)
	}

	log.Infof("Created r tree index with %d length in %s", index.tree.Size(), time.Since(then))

	return index
}

func (store *RTreeIndex) AddShape(shape geometry.Shape) {
	rTreeObject, err := newRTreeObjectFromShape(shape)

	if err != nil {
		log.WithError(err).Warnf("Could not add shape %#v, boundingbox %#v", shape, shape.GetBoundingBox())
		return
	}

	store.treeObjects = append(store.treeObjects, rTreeObject)
	store.tree.Insert(rTreeObject)
}

// AddShapes ...
func (store *RTreeIndex) AddShapes(shapes []geometry.Shape) {
	for _, shape := range shapes {
		store.AddShape(shape)
	}
}

// GetShapeByName ...
func (store *RTreeIndex) GetShapeByName(shapeName string) (geometry.Shape, error) {
	for _, treeObject := range store.treeObjects {
		if treeObject.shape.GetName() == shapeName {
			return treeObject.shape, nil
		}
	}

	return nil, errors.New("No shape found with name " + shapeName)
}

// RemoveShapeByName ...
func (store *RTreeIndex) RemoveShapeByName(shapeName string) (geometry.Shape, error) {
	for _, treeObject := range store.treeObjects {
		if treeObject.shape.GetName() == shapeName {
			removed := store.tree.Delete(treeObject)

			if !removed {
				return treeObject.shape, fmt.Errorf("Could not find shape to remove")
			}

			return treeObject.shape, nil
		}
	}

	return &geometry.Circle{}, errors.New("No shape found with name " + shapeName)
}

// FindShapesWhichContainsPoint ...
func (store *RTreeIndex) FindShapesWhichContainsPoint(point geometry.Point) []geometry.Shape {
	var matchingShapes []geometry.Shape = make([]geometry.Shape, 0)

	// RTrees are not really meant for lookup on points, so we create a minimal bounding box
	// and feed it to SearchIntersect
	searchRect, err := rtreego.NewRect(rtreego.Point{point.X, point.Y}, []float64{0.00001, 0.00001})
	if err != nil {
		log.WithError(err).Warnf("Failed to create new rect")
		return matchingShapes
	}

	results := store.tree.SearchIntersect(
		searchRect,
	)

	for _, result := range results {
		if treeObject, ok := result.(*rTreeObject); ok && treeObject.shape.PointInside(&point) {
			matchingShapes = append(matchingShapes, treeObject.shape)
		}
	}

	return matchingShapes
}

// FindShapesWhichContainsShape ...
func (store *RTreeIndex) FindShapesWhichContainsShape(shape geometry.Shape) []geometry.Shape {
	var matchingShapes []geometry.Shape = make([]geometry.Shape, 0)

	// RTrees are not really meant for lookup on points, so we create a minimal bounding box
	// and feed it to SearchIntersect
	searchRect, err := newRectFromBoundingBox(shape.GetBoundingBox())

	if err != nil {
		log.WithError(err).Warnf("Failed to create new rect")
		return matchingShapes
	}

	results := store.tree.SearchIntersect(searchRect)

	for _, result := range results {
		if treeObject, ok := result.(*rTreeObject); ok && treeObject.shape.ShapeInside(shape) {
			matchingShapes = append(matchingShapes, treeObject.shape)
		}
	}

	return matchingShapes

}

func newRTreeObjectFromShape(shape geometry.Shape) (*rTreeObject, error) {
	rect, err := newRectFromBoundingBox(shape.GetBoundingBox())

	if err != nil {
		return &rTreeObject{}, err
	}

	return &rTreeObject{
		shape: shape,
		rect:  rect,
	}, nil
}

func newRectFromBoundingBox(boundingBox geometry.BoundingBox) (*rtreego.Rect, error) {
	return rtreego.NewRect(
		rtreego.Point{boundingBox.MinX, boundingBox.MinY},
		[]float64{boundingBox.MaxX - boundingBox.MinX, boundingBox.MaxY - boundingBox.MinY},
	)
}
