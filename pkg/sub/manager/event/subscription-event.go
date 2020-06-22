package event

import (
	"github.com/eesrc/geo/pkg/model"
)

// SubscriptionEvent wraps an event with a identifier to help
type SubscriptionEvent struct {
	Type EventType                `json:"type"`
	Data SubscriptionEventDetails `json:"data"`
}

// EventType returns the Subscription event type
func (subscriptionEvent *SubscriptionEvent) EventType() EventType {
	return subscriptionEvent.Type
}

// SubscriptionEventDetails is a struct to be used for events
// triggered by a subscription
type SubscriptionEventDetails struct {
	SubscriptionID int64          `json:"subscriptionId"`
	Position       model.Position `json:"position"`
	Details        TriggerDetails `json:"details"`
}

// TriggerDetails contains information about why the subscription triggered
// along with IDs to shapecollection and shape
type TriggerDetails struct {
	Movements         []string `json:"movements"`
	ShapecollectionID int64    `json:"shapecollectionId"`
	ShapeID           int64    `json:"shapeId"`
}

// NewSubscriptionEvent returns a new wrapped PositionEvent
func NewSubscriptionEvent(subscriptionID int64, position model.Position, triggerDetails TriggerDetails) *SubscriptionEvent {
	return &SubscriptionEvent{
		Type: Subscription,
		Data: SubscriptionEventDetails{
			SubscriptionID: subscriptionID,
			Position:       position,
			Details:        triggerDetails,
		},
	}
}
