package service

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/model"
)

// Collection is the API representation of a collection
type Collection struct {
	ID          int64  `json:"id"`
	TeamID      int64  `json:"teamId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToModel creates a storage model from the API representation
func (collection *Collection) ToModel() *model.Collection {
	return &model.Collection{
		ID:          collection.ID,
		TeamID:      collection.TeamID,
		Name:        collection.Name,
		Description: collection.Description,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (collection *Collection) MarshalJSON() ([]byte, error) {
	return json.Marshal(*collection)
}

// NewCollectionFromModel creates a HTTP representation of a model collection
func NewCollectionFromModel(collectionModel *model.Collection) *Collection {
	return &Collection{
		ID:          collectionModel.ID,
		TeamID:      collectionModel.TeamID,
		Name:        collectionModel.Name,
		Description: collectionModel.Description,
	}
}
