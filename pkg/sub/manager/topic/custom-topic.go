package topic

// NatsCustomTopic is a Collection specific topic, wrapping a generic topic
type NatsCustomTopic struct {
	topicString string
	subject     SubjectType
	events      EventType
}

// Subject returns AllSubjects as default
func (topic *NatsCustomTopic) Subject() SubjectType {
	return topic.subject
}

// Event returns AllEvents as default
func (topic *NatsCustomTopic) Event() EventType {
	return topic.events
}

// SetEvents sets the custom topic events
func (topic *NatsCustomTopic) SetEvents(eventType EventType) {
	topic.events = eventType
}

// TopicString returns a the given topic string
func (topic *NatsCustomTopic) TopicString() string {
	return topic.topicString
}

// NewCustomTopic creates a default NatsCustomTopic with a topic string
func NewCustomTopic(topicString string) *NatsCustomTopic {
	return &NatsCustomTopic{
		topicString: topicString,
		subject:     AllSubjects,
		events:      AllEvents,
	}
}
