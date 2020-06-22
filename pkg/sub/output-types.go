package sub

// OutputType Type of output for a subscription
type OutputType string

const (
	// Console print console output. Debug only.
	Console OutputType = "console"
	// SMS sets up an SMS output type
	SMS OutputType = "sms"
	// Webhook sets up a webhook output type
	Webhook OutputType = "webhook"
	// WebSocket enables simple streaming through a subscription
	WebSocket OutputType = "websocket"
)

// ValidOutputTypes is a list of valid OutputTypes for the subscriptions
var ValidOutputTypes = []OutputType{SMS, Webhook, WebSocket}
