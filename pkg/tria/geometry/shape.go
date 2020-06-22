package geometry

import (
	"database/sql/driver"

	"github.com/eesrc/geo/pkg/serializing"
)

// ShapeProperties represents arbitrary string properties connected to a shape
type ShapeProperties map[string]interface{}

// Value implements SQL value driver
func (shapeProperties ShapeProperties) Value() (driver.Value, error) {
	return serializing.ValueJSON(shapeProperties)
}

// Scan implements SQL scan driver
func (shapeProperties *ShapeProperties) Scan(src interface{}) error {
	return serializing.ScanJSON(shapeProperties, src)
}

// Shape is a interface which is an abstract way of checking if a point exists within it
type Shape interface {
	GetName() string
	Type() string

	GetID() int64
	SetID(int64)

	GetProperties() ShapeProperties
	SetProperties(shapeProperties ShapeProperties)

	GetBoundingBox() BoundingBox

	PointInside(point *Point) bool
	PointInsideBoundingBox(point *Point) bool

	ShapeInside(shape Shape) bool
}
