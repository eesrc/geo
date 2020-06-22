package service

import (
	"github.com/eesrc/geo/pkg/sub/manager/event"
)

// SubscriptionEvent wraps an event with a type to help distinguish the event
type SubscriptionEvent struct {
	Type event.EventType          `json:"type"`
	Data SubscriptionEventDetails `json:"data"`
}

// EventType returns the Position event type
func (subscriptionEvent *SubscriptionEvent) EventType() event.EventType {
	return subscriptionEvent.Type
}

// SubscriptionEventDetails is a struct to be used for events
// triggered by a new position
type SubscriptionEventDetails struct {
	SubscriptionID int64          `json:"subscriptionId"`
	Position       *Position      `json:"position"`
	Details        TriggerDetails `json:"details"`
}

// TriggerDetails contains information about why the subscription triggered
// along with IDs to shapecollection and shape
type TriggerDetails struct {
	Movements         []string `json:"movements"`
	ShapecollectionID int64    `json:"shapeCollectionId"`
	ShapeID           int64    `json:"shapeId"`
}

// NewSubscriptionEventFromModel returns a new service SubscriptionEvent from a model
func NewSubscriptionEventFromModel(subscriptionEvent *event.SubscriptionEvent) *SubscriptionEvent {
	return &SubscriptionEvent{
		Type: subscriptionEvent.Type,
		Data: SubscriptionEventDetails{
			SubscriptionID: subscriptionEvent.Data.SubscriptionID,
			Position:       NewPositionFromModel(&subscriptionEvent.Data.Position),
			Details: TriggerDetails{
				Movements:         subscriptionEvent.Data.Details.Movements,
				ShapecollectionID: subscriptionEvent.Data.Details.ShapecollectionID,
				ShapeID:           subscriptionEvent.Data.Details.ShapeID,
			},
		},
	}
}
