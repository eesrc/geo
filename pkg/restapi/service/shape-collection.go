package service

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/model"
)

// ShapeCollection is the API representation of a shape collection
type ShapeCollection struct {
	ID          int64  `json:"id"`
	TeamID      int64  `json:"teamId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToModel creates a storage model from the API representation
func (shapeCollection *ShapeCollection) ToModel() *model.ShapeCollection {
	return &model.ShapeCollection{
		ID:          shapeCollection.ID,
		TeamID:      shapeCollection.TeamID,
		Name:        shapeCollection.Name,
		Description: shapeCollection.Description,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (shapeCollection *ShapeCollection) MarshalJSON() ([]byte, error) {
	return json.Marshal(*shapeCollection)
}

// NewShapeCollectionFromModel creates a HTTP representation of a model shape collection
func NewShapeCollectionFromModel(shapeCollectionModel *model.ShapeCollection) *ShapeCollection {
	return &ShapeCollection{
		ID:          shapeCollectionModel.ID,
		TeamID:      shapeCollectionModel.TeamID,
		Name:        shapeCollectionModel.Name,
		Description: shapeCollectionModel.Description,
	}
}
