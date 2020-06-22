package service

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/sub"
	"github.com/eesrc/geo/pkg/sub/output"
)

// Subscription is a representation of subscription for a collection or
// tracker towards a output
type Subscription struct {
	ID                int64           `json:"id"`
	TeamID            *int64          `json:"teamId"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Active            bool            `json:"active"`
	Output            OutputEntry     `json:"output"`
	TriggerCriteria   TriggerCriteria `json:"triggerCriteria"`
	ShapeCollectionID *int64          `json:"shapeCollectionId"`
	Trackable         TrackableEntry  `json:"trackable"`
}

// OutputEntry contains information about the subscriptions output
type OutputEntry struct {
	Type   sub.OutputType `json:"type"`
	Config output.Config  `json:"config"`
}

// TrackableEntry contains information about the subscriptions trackable
type TrackableEntry struct {
	Type sub.TrackableType `json:"type"`
	ID   *int64            `json:"id"`
}

// TriggerCriteria contains information about the
type TriggerCriteria struct {
	TriggerTypes sub.MovementList   `json:"triggerTypes"`
	Confidence   sub.ConfidenceList `json:"confidence"`
}

// ToModel creates a storage model from the API representation
func (subscription *Subscription) ToModel() *model.Subscription {
	return &model.Subscription{
		ID:                subscription.ID,
		TeamID:            *subscription.TeamID,
		Name:              subscription.Name,
		Description:       subscription.Description,
		Active:            subscription.Active,
		Output:            string(subscription.Output.Type),
		OutputConfig:      model.OutputConfig(subscription.Output.Config),
		Types:             subscription.TriggerCriteria.TriggerTypes.ToModel(),
		Confidences:       subscription.TriggerCriteria.Confidence.ToModel(),
		ShapeCollectionID: *subscription.ShapeCollectionID,
		TrackableType:     string(subscription.Trackable.Type),
		TrackableID:       *subscription.Trackable.ID,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (subscription *Subscription) MarshalJSON() ([]byte, error) {
	return json.Marshal(*subscription)
}

// NewSubscriptionFromModel creates a HTTP representation of a model subscription
func NewSubscriptionFromModel(subscriptionModel *model.Subscription) *Subscription {
	shapeID := subscriptionModel.ShapeCollectionID
	trackableID := subscriptionModel.TrackableID

	return &Subscription{
		ID:          subscriptionModel.ID,
		TeamID:      &subscriptionModel.TeamID,
		Name:        subscriptionModel.Name,
		Description: subscriptionModel.Description,
		Active:      subscriptionModel.Active,
		Output: OutputEntry{
			Type:   sub.OutputType(subscriptionModel.Output),
			Config: output.Config(subscriptionModel.OutputConfig),
		},
		TriggerCriteria: TriggerCriteria{
			TriggerTypes: sub.NewMovementTypeFromModel(subscriptionModel.Types),
			Confidence:   sub.NewConfidenceListFromModel(subscriptionModel.Confidences),
		},
		ShapeCollectionID: &shapeID,
		Trackable: TrackableEntry{
			Type: sub.TrackableType(subscriptionModel.TrackableType),
			ID:   &trackableID,
		},
	}
}

// NewSubscription returns a new instance of Subscription
func NewSubscription() Subscription {
	return Subscription{
		Active: false,
		Output: OutputEntry{
			Type:   sub.Webhook,
			Config: output.Config{},
		},
		TriggerCriteria: TriggerCriteria{},
		Trackable: TrackableEntry{
			Type: sub.Collection,
		},
	}
}
