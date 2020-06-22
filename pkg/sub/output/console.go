package output

import (
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
)

// ConsoleOutput is a dummy output which simply prints to console whenever a subscription matches.
type ConsoleOutput struct {
	name            string
	nReceived       int32
	terminate       chan bool
	mutex           sync.Mutex
	geoSubscription GeoSubscription
	config          Config
	eventCallback   func(topic topic.Topic, event event.PublishableEvent)
}

func (consoleOutput *ConsoleOutput) messageReader(receiver <-chan interface{}) {
	for {
		select {
		case <-consoleOutput.terminate:
			return
		case msg, ok := <-receiver:
			if !ok {
				return
			}
			consolePayloads := make([]outputPayload, 0)

			for {
				outputPayload, err := consoleOutput.geoSubscription.GetOutputPayloadFromEvent(msg)

				if err != nil {
					log.Error("Something went wrong when trying to get outputPayload", err)
					continue
				}

				if len(outputPayload.movements) > 0 {
					consolePayloads = append(consolePayloads, outputPayload)
				}

				if len(receiver) == 0 {
					break
				}
				msg = <-receiver
			}

			atomic.AddInt32(&consoleOutput.nReceived, 1)
			consoleOutput.printToConsole(consolePayloads)
		}
	}
}

// Start initiate start of the console outpt
func (consoleOutput *ConsoleOutput) Start(config Config, message <-chan interface{}) {
	consoleOutput.mutex.Lock()
	consoleOutput.config = config
	defer consoleOutput.mutex.Unlock()

	go consoleOutput.messageReader(message)
}

// Stop initiate a stop of the console output
func (consoleOutput *ConsoleOutput) Stop(timeout time.Duration) {
	select {
	case consoleOutput.terminate <- true:
	default:
	}
}

// Validate validates a console configuration
func (consoleOutput *ConsoleOutput) Validate(config Config) error {
	return nil
}

func (consoleOutput *ConsoleOutput) printToConsole(payloads []outputPayload) {
	for _, payload := range payloads {
		for _, movement := range payload.movements {
			if consoleOutput.geoSubscription.ContainsAnyMovements(movement.lastMovements) {
				log.Infof("Console subscription %d with sub to movements %v. Shape with ID %v. Movements triggering subscription %v",
					consoleOutput.geoSubscription.Subscription.ID,
					consoleOutput.geoSubscription.Subscription.Types,
					movement.shapeID,
					movement.lastMovements,
				)

				// Publish trigger event
				consoleOutput.eventCallback(
					topic.NewEntityTopic(topic.Subscription, consoleOutput.geoSubscription.Subscription.ID, topic.TriggerEvents),
					event.NewSubscriptionEvent(
						consoleOutput.geoSubscription.Subscription.ID,
						payload.position,
						event.TriggerDetails{
							Movements:         movement.lastMovements.ToStringSlice(),
							ShapecollectionID: consoleOutput.geoSubscription.Subscription.ShapeCollectionID,
							ShapeID:           movement.shapeID,
						},
					),
				)
			}
		}
	}
}

// NewConsoleOutput creates a new console output
func NewConsoleOutput(geoSubscription GeoSubscription, eventCallback func(topic topic.Topic, event event.PublishableEvent)) *ConsoleOutput {
	config := Config(geoSubscription.Subscription.OutputConfig)

	name := config.GetStringWithDefault("name", "Console output")

	return &ConsoleOutput{
		name:            name,
		geoSubscription: geoSubscription,
		eventCallback:   eventCallback,
	}
}
