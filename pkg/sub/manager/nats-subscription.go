package manager

import (
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	nats "github.com/nats-io/nats.go"
)

// NatsSubscription is a subscription with NATS as backend, creating a subscription and piping into a channel
type NatsSubscription struct {
	Subscription *nats.Subscription
	Chan         chan interface{}
}

// NewNatsSubscription initializes a Susbcription on given connection and pipes the data into
// a local channel
func NewNatsSubscription(subscriptionTopic topic.Topic, connection *nats.EncodedConn) (*NatsSubscription, error) {
	channel := make(chan interface{})

	subscription, err := connection.Subscribe(subscriptionTopic.TopicString(), func(message *nats.Msg) {
		channel <- message.Data
	})
	if err != nil {
		return &NatsSubscription{}, err
	}

	return &NatsSubscription{
		Subscription: subscription,
		Chan:         channel,
	}, nil
}

// Unsubscribe cleans up the NATS-subscription and closes the channel
func (natsSub *NatsSubscription) Unsubscribe() error {
	if !natsSub.Subscription.IsValid() {
		return nil
	}

	err := natsSub.Subscription.Unsubscribe()
	close(natsSub.Chan)
	return err
}

// GetChan returns a channel for retrieving data
func (natsSub *NatsSubscription) GetChan() <-chan interface{} {
	return natsSub.Chan
}
