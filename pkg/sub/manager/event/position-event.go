package event

import (
	"github.com/eesrc/geo/pkg/model"
)

// PositionEvent wraps an event with a type identifier to help distinguish the event from other events
type PositionEvent struct {
	Type EventType            `json:"type"`
	Data PositionEventDetails `json:"data"`
}

// EventType returns the Position event type
func (positionEvent *PositionEvent) EventType() EventType {
	return positionEvent.Type
}

// PositionEventDetails is a struct to be used for events
// triggered by a new position
type PositionEventDetails struct {
	CollectionID int64          `json:"collectionId"`
	TrackerID    int64          `json:"trackerId"`
	Position     model.Position `json:"position"`
}

// NewPositionEvent returns a new wrapped PositionEvent
func NewPositionEvent(collectionID int64, position model.Position) *PositionEvent {
	return &PositionEvent{
		Type: Position,
		Data: PositionEventDetails{
			CollectionID: collectionID,
			TrackerID:    position.TrackerID,
			Position:     position,
		},
	}
}
