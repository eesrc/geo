package service

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/model"
)

// Tracker is the API representation of a tracker
type Tracker struct {
	ID           int64  `json:"id"`
	CollectionID int64  `json:"collectionId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

// ToModel creates a storage model from the API representation
func (tracker *Tracker) ToModel() *model.Tracker {
	return &model.Tracker{
		ID:           tracker.ID,
		CollectionID: tracker.CollectionID,
		Name:         tracker.Name,
		Description:  tracker.Description,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (tracker *Tracker) MarshalJSON() ([]byte, error) {
	return json.Marshal(*tracker)
}

// NewTrackerFromModel creates a HTTP representation of a model tracker
func NewTrackerFromModel(trackerModel *model.Tracker) *Tracker {
	return &Tracker{
		ID:           trackerModel.ID,
		CollectionID: trackerModel.CollectionID,
		Name:         trackerModel.Name,
		Description:  trackerModel.Description,
	}
}
