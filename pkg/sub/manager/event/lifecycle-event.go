package event

// LifeCycleEventType defines a lifecycle type
type LifeCycleEventType string

// LifeCycleEvent wraps an event with a type identifier to help distinguish the event from other events
type LifeCycleEvent struct {
	Type EventType             `json:"type"`
	Data LifecycleEventDetails `json:"data"`
}

// EventType returns the Lifecycle event type
func (lifecycleEvent *LifeCycleEvent) EventType() EventType {
	return lifecycleEvent.Type
}

const (
	// CreatedEvent is when an entity is created
	CreatedEvent LifeCycleEventType = "created"
	// UpdatedEvent is when an entity is updated
	UpdatedEvent LifeCycleEventType = "updated"
	// DeletedEvent is when an entity is deleted
	DeletedEvent LifeCycleEventType = "deleted"
)

// LifecycleEventDetails is a struct to be used for events
// triggered by a lifecycle change
type LifecycleEventDetails struct {
	Type       LifeCycleEventType `json:"type"`
	EntityType EntityType         `json:"entityType"`
	EntityID   int64              `json:"entityId"`
}

// NewLifecycleEvent returns a new wrapped LifecycleEvent
func NewLifecycleEvent(eventType LifeCycleEventType, entityType EntityType, id int64) *LifeCycleEvent {
	return &LifeCycleEvent{
		Type: LifeCycle,
		Data: LifecycleEventDetails{
			Type:       eventType,
			EntityType: entityType,
			EntityID:   id,
		},
	}
}
