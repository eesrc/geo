package event

// PublishableEvent helps us have type safety on events that are publishable
type PublishableEvent interface {
	EventType() EventType
}
