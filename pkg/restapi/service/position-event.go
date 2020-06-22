package service

import "github.com/eesrc/geo/pkg/sub/manager/event"

// PositionEvent wraps an event with a type to help distinguish the event
type PositionEvent struct {
	Type event.EventType      `json:"type"`
	Data PositionEventDetails `json:"data"`
}

// EventType returns the Position event type
func (positionEvent *PositionEvent) EventType() event.EventType {
	return positionEvent.Type
}

// PositionEventDetails is a struct to be used for events
// triggered by a new position
type PositionEventDetails struct {
	CollectionID int64     `json:"collectionId"`
	TrackerID    int64     `json:"trackerId"`
	Position     *Position `json:"position"`
}

// NewPositionEventFromModel returns a new service PositionEvent from a model
func NewPositionEventFromModel(positionEventModel *event.PositionEvent) *PositionEvent {
	return &PositionEvent{
		Type: positionEventModel.Type,
		Data: PositionEventDetails{
			CollectionID: positionEventModel.Data.CollectionID,
			TrackerID:    positionEventModel.Data.TrackerID,
			Position:     NewPositionFromModel(&positionEventModel.Data.Position),
		},
	}
}
