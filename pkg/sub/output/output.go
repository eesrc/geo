package output

import (
	"fmt"
	"time"

	"github.com/eesrc/geo/pkg/sub"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
)

// Output is the generic interface to create a working output for subscriptions
type Output interface {
	// Start launches the output. This is non blocking, ie if the output
	// must connect to a remote server or perform some sort initialization it
	// will do so in a separate goroutine. The output will stop automatically
	// when the message channel is closed. If the message channel is closed
	// the output should attempt empty any remaining messages in the queue.
	Start(config Config, message <-chan interface{})

	// Stop halts the output. Any buffered messages that can't be sent during
	// the timeout will be discarded by the output. When the Stop call returns
	// the output has stopped.
	Stop(timeout time.Duration)
}

// NewOutput initializes a new output
func NewOutput(geoSubscription GeoSubscription, eventCallback func(topic topic.Topic, event event.PublishableEvent)) (Output, error) {
	var outputType sub.OutputType = sub.OutputType(geoSubscription.Subscription.Output)

	switch outputType {
	case sub.SMS:
		return NewConsoleOutput(geoSubscription, eventCallback), nil
	case sub.Webhook:
		return NewConsoleOutput(geoSubscription, eventCallback), nil
	case sub.WebSocket:
		return NewWebsocketOutput(geoSubscription, eventCallback), nil
	}

	return nil, fmt.Errorf("Could not find a output with type '%s'", geoSubscription.Subscription.Output)
}
