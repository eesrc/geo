package manager

import (
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/eesrc/geo/pkg/sub/output"
)

// Manager is responsible for keeping track of events and event subscriptions along with handling
// GeoSubscriptions which contains geo indexes for lookup and calculations
type Manager interface {
	// Load loads GeoSubscriptions from backend store and launches the ones that aren't
	// up and running yet. The Load call might be performed multiple times to
	// update the list.
	Refresh([]output.GeoSubscription)

	// Update refreshes the GeoSubscription. If it isn't launched yet it will be
	// launched. If it is already running the new configuration will be applied.
	Update(output.GeoSubscription) error

	// Stop stops a single subscription, typically if they have been deleted.
	Stop(subscriptionID int64) error

	// Shutdown shuts down all of the running subscriptions.
	Shutdown()

	// Get returns the GeoSubscription. If the subscription isn't running or is unknown it
	// will return an error.
	Get(SubscriptionID int64) (output.GeoSubscription, error)

	// Publish publishes an event to the event bus on the given topic. If there's no
	// subscriptions subscribing to the topic it will be discarded.
	Publish(topic topic.Topic, event event.PublishableEvent)

	// Subscribe subscribes to a topic
	Subscribe(topic.Topic) (Subscription, error)
}
