package topic

import (
	"fmt"

	"github.com/eesrc/geo/pkg/model"
)

// SubjectType is the kind of subjects you're subscribing to
type SubjectType string

const (
	// AllSubjects represents all type of event subjects
	AllSubjects SubjectType = "*"
	// Team represents all events with subject teams
	Team SubjectType = "teams"
	// Collection represents all events with subject collection
	Collection SubjectType = "collections"
	// Tracker represents all events with subject tracker
	Tracker SubjectType = "trackers"
	// Subscription represents all events with subject subscriptions
	Subscription SubjectType = "subscriptions"
	// ShapeCollections represents all events with subject shape collections
	ShapeCollections SubjectType = "shapecollections"
	// Token represents all events with subject tokens
	Token SubjectType = "tokens"
)

// EventType is the kind of event you're subscribing to
type EventType string

const (
	// AllEvents represents all types of events
	AllEvents EventType = "*"
	// DataEvents represents all data type events
	DataEvents EventType = "data"
	// LifecycleEvents represents all lifecycle events
	LifecycleEvents EventType = "lifecycle"
	// TriggerEvents represents events that are triggered, often by subscriptions
	TriggerEvents EventType = "trigger"
)

// GetTopicString returns a topic string based on given SubjectType, EventType and ID
func GetTopicString(subject SubjectType, event EventType, id int64) string {
	return fmt.Sprintf("%s.%d.%s", subject, id, event)
}

// GetTopicFromSubscription returns a Topic based on subscription and EventType
func GetTopicFromSubscription(subscription model.Subscription, event EventType) Topic {
	if subscription.TrackableType == string(Collection) {
		return NewEntityTopic(Collection, subscription.TrackableID, event)
	}

	if subscription.TrackableType == string(Tracker) {
		return NewEntityTopic(Tracker, subscription.TrackableID, event)
	}

	// TODO: Is it correct to default to a collection topic?
	return NewEntityTopic(Collection, subscription.TrackableID, event)
}
