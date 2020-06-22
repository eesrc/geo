package topic

// Topic represents a subscribable topic
type Topic interface {
	Subject() SubjectType
	Event() EventType

	TopicString() string
}
