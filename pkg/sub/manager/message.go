package manager

// Message is an intermediate message wrapping the payload with a type
// so you can correctly get the right interface
type Message interface {
	Type() string
	Payload() []byte
}
