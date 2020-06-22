package service

import "github.com/eesrc/geo/pkg/sub/manager/event"

// LifecycleEvent wraps an event with a type to help distinguish the event
type LifecycleEvent struct {
	Type event.EventType       `json:"type"`
	Data LifecycleEventDetails `json:"data"`
}

// LifeCycleEvent wraps an event with a type identifier to help distinguish the event from other events
type LifeCycleEvent struct {
	Type event.EventType       `json:"type"`
	Data LifecycleEventDetails `json:"data"`
}

// LifecycleEventDetails is a struct to be used for events
// triggered by a lifecycle change
type LifecycleEventDetails struct {
	Type       event.LifeCycleEventType `json:"type"`
	EntityType event.EntityType         `json:"entityType"`
	EntityID   int64                    `json:"entityId"`
}

// NewLifecycleEventFromModel returns a new service LifecycleEvent from a model
func NewLifecycleEventFromModel(lifecycleEventModel *event.LifeCycleEvent) *LifecycleEvent {
	return &LifecycleEvent{
		Type: lifecycleEventModel.Type,
		Data: LifecycleEventDetails{
			Type:       lifecycleEventModel.Data.Type,
			EntityType: lifecycleEventModel.Data.EntityType,
			EntityID:   lifecycleEventModel.Data.EntityID,
		},
	}
}
