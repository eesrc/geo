package service

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/tria/geometry"
)

// Shape is the API representation of a shape
type Shape struct {
	ID                int64                    `json:"id"`
	ShapeCollectionID int64                    `json:"shapeCollectionId"`
	Name              string                   `json:"name"`
	Properties        geometry.ShapeProperties `json:"properties"`
}

// ToModel creates a storage model from the API representation
func (shape *Shape) ToModel() *model.Shape {
	shapeProperties := shape.Properties

	// Radius is a special case where we assume m and reduce it to a simplified WGS84 compliant value
	if radius, ok := shapeProperties["radius"]; ok {
		if radiusNumber, ok := radius.(float64); ok {
			shapeProperties["radius"] = radiusNumber / 111111
		} else {
			delete(shapeProperties, "radius")
		}
	}

	return &model.Shape{
		ID:                shape.ID,
		ShapeCollectionID: shape.ShapeCollectionID,
		Name:              shape.Name,
		Properties:        shapeProperties,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (shape *Shape) MarshalJSON() ([]byte, error) {
	return json.Marshal(*shape)
}

// NewShapeFromModel creates a HTTP representation of a model shape
func NewShapeFromModel(shapeModel *model.Shape) *Shape {
	properties := shapeModel.Properties
	if properties == nil {
		properties = geometry.ShapeProperties{}
	}

	// Radius is a special case where we assume WGS84 compliant value and multiply the value to represent m
	if radius, ok := properties["radius"]; ok {
		if radiusNumber, ok := radius.(float64); ok {
			properties["radius"] = radiusNumber * 111111
		}
	}

	return &Shape{
		ID:                shapeModel.ID,
		ShapeCollectionID: shapeModel.ShapeCollectionID,
		Name:              shapeModel.Name,
		Properties:        properties,
	}
}
