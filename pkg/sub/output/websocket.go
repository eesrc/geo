package output

import (
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
)

// WebsocketOutput is a simple output to enable the subscription on the /stream-endpoint of a subscription
type WebsocketOutput struct {
	name            string
	nReceived       int32
	terminate       chan bool
	mutex           sync.Mutex
	geoSubscription GeoSubscription
	config          Config
	eventCallback   func(topic topic.Topic, event event.PublishableEvent)
}

func (websocketOutput *WebsocketOutput) messageReader(receiver <-chan interface{}) {
	for {
		select {
		case <-websocketOutput.terminate:
			return
		case msg, ok := <-receiver:
			if !ok {
				return
			}
			eventPayloads := make([]outputPayload, 0)

			for {
				outputPayload, err := websocketOutput.geoSubscription.GetOutputPayloadFromEvent(msg)

				if err != nil {
					log.Error("Something went wrong when trying to get outputPayload", err)
					continue
				}

				if len(outputPayload.movements) > 0 {
					eventPayloads = append(eventPayloads, outputPayload)
				}

				if len(receiver) == 0 {
					break
				}
				msg = <-receiver
			}

			atomic.AddInt32(&websocketOutput.nReceived, 1)
			websocketOutput.publishToWebsocket(eventPayloads)
		}
	}
}

// Start initiate start of the console outpt
func (websocketOutput *WebsocketOutput) Start(config Config, message <-chan interface{}) {
	websocketOutput.mutex.Lock()
	websocketOutput.config = config
	defer websocketOutput.mutex.Unlock()

	go websocketOutput.messageReader(message)
}

// Stop initiate a stop of the console output
func (websocketOutput *WebsocketOutput) Stop(timeout time.Duration) {
	select {
	case websocketOutput.terminate <- true:
	default:
	}
}

// Validate validates a console configuration
func (websocketOutput *WebsocketOutput) Validate(config Config) error {
	return nil
}

func (websocketOutput *WebsocketOutput) publishToWebsocket(payloads []outputPayload) {
	for _, payload := range payloads {
		for _, movement := range payload.movements {
			if websocketOutput.geoSubscription.ContainsAnyMovements(movement.lastMovements) {
				// Publish trigger event
				websocketOutput.eventCallback(
					topic.NewEntityTopic(topic.Subscription, websocketOutput.geoSubscription.Subscription.ID, topic.TriggerEvents),
					event.NewSubscriptionEvent(
						websocketOutput.geoSubscription.Subscription.ID,
						payload.position,
						event.TriggerDetails{
							Movements:         movement.lastMovements.ToStringSlice(),
							ShapecollectionID: websocketOutput.geoSubscription.Subscription.ShapeCollectionID,
							ShapeID:           movement.shapeID,
						},
					),
				)
			}
		}
	}
}

// NewWebsocketOutput creates a new console output
func NewWebsocketOutput(geoSubscription GeoSubscription, eventCallback func(topic topic.Topic, event event.PublishableEvent)) *WebsocketOutput {
	config := Config(geoSubscription.Subscription.OutputConfig)

	name := config.GetStringWithDefault("name", "Websocket output")

	return &WebsocketOutput{
		name:            name,
		geoSubscription: geoSubscription,
		eventCallback:   eventCallback,
	}
}
