package topic

import "fmt"

// NatsEntityTopic is a Collection specific topic, wrapping a generic topic
type NatsEntityTopic struct {
	entityID   int64
	eventType  EventType
	entityType SubjectType
}

// Subject returns the configured subject
func (topic *NatsEntityTopic) Subject() SubjectType {
	return topic.entityType
}

// Event returns the configured event type
func (topic *NatsEntityTopic) Event() EventType {
	return topic.eventType
}

// TopicString returns a formatted topic string
func (topic *NatsEntityTopic) TopicString() string {
	return fmt.Sprintf("%s.%d.%s", topic.Subject(), topic.entityID, topic.Event())
}

// NewEntityTopic creates a default NatsEntityTopic with given id and eventType
func NewEntityTopic(entityType SubjectType, ID int64, eventType EventType) *NatsEntityTopic {
	return &NatsEntityTopic{
		entityType: entityType,
		entityID:   ID,
		eventType:  eventType,
	}
}
