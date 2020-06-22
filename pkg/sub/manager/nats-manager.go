package manager

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/eesrc/geo/pkg/sub/output"
	stand "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats-streaming-server/stores"
	nats "github.com/nats-io/nats.go"
)

// natsManager is a manager running on the local instance. It will only keep
// track of outputs launched locally.
type natsManager struct {
	running    map[int64]subscriberEntry
	publisher  *nats.EncodedConn
	natsServer *stand.StanServer
	natsConn   *nats.Conn
	mutex      *sync.Mutex
}

type subscriberEntry struct {
	sub             Subscription
	output          output.Output
	geoSubscription output.GeoSubscription
}

type NATSManagerConfig struct {
	Logging   bool   `param:"desc=Whether the NATS server should show debug messages;default=false"`
	StoreType string `param:"desc=The type of store used for NATS;options=memory,file;default=memory"`

	// File-specific config
	FileStoreDir string `param:"desc=The directory where to store the NATS subscriptions (if store-type is 'file');default=nats"`

	// NATS connection
	Host string `param:"desc=The host url;default=localhost"`
	Port int    `param:"desc=;default=4222"`
}

// NewNatsManager creates a new manager
func NewNatsManager(config NATSManagerConfig) Manager {
	server, err := createNATSStreamingServer(config)

	if err != nil {
		log.Fatalf("Failed to start NATS server: %v", err)
	}

	// Connect to the newly started NATS server
	natsConnection, err := nats.Connect(fmt.Sprintf("%s:%d", config.Host, config.Port),
		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
			log.Infof("Known servers: %v\n", nc.Servers())
			log.Infof("Discovered servers: %v\n", nc.DiscoveredServers())
		}),
	)

	if err != nil {
		log.Fatalf("Failed to connect to NATS server %v", err)
	}

	log.Infof("Connected to NATS running on '%s'", natsConnection.ConnectedAddr())

	// Create an encoded connection which will automatically encode payloads as JSON
	encodedConnection, err := nats.NewEncodedConn(natsConnection, nats.JSON_ENCODER)

	if err != nil {
		log.Fatal("Failed to initialize nats", err)
	}

	return &natsManager{
		running:    make(map[int64]subscriberEntry),
		natsServer: server,
		natsConn:   natsConnection,
		publisher:  encodedConnection,
		mutex:      &sync.Mutex{},
	}
}

func (manager *natsManager) Refresh(geoSubscriptions []output.GeoSubscription) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for _, geoSubscription := range geoSubscriptions {
		_, exists := manager.running[geoSubscription.Subscription.ID]
		if exists {
			continue
		}
		if !geoSubscription.Subscription.Active {
			continue
		}
		newOutput, err := output.NewOutput(geoSubscription, manager.Publish)
		if err != nil {
			log.WithError(err).Errorf("Unable to launch subscription with ID %d. Ignoring", geoSubscription.Subscription.ID)
			continue
		}

		// Subscribing to generated topic based on Subscription
		sub, err := manager.Subscribe(topic.GetTopicFromSubscription(geoSubscription.Subscription, topic.DataEvents))
		if err != nil {
			log.WithError(err).Errorf(
				"Unable to subscribe to topic '%s'",
				topic.GetTopicFromSubscription(geoSubscription.Subscription, topic.DataEvents).TopicString(),
			)
			continue
		}

		// Start output and store in running output map
		newOutput.Start(output.Config(geoSubscription.Subscription.OutputConfig), sub.GetChan())
		manager.running[geoSubscription.Subscription.ID] = subscriberEntry{sub: sub, output: newOutput, geoSubscription: geoSubscription}
	}
}

func (manager *natsManager) Update(geoSubscription output.GeoSubscription) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	subscription := geoSubscription.Subscription

	v, exists := manager.running[subscription.ID]
	if exists {
		unsubscribeSubscription(v.sub)
		v.output.Stop(stopTimeout)
	}

	if !subscription.Active {
		return nil
	}

	newOutput, err := output.NewOutput(geoSubscription, manager.Publish)
	if err != nil {
		return err
	}

	sub, err := manager.Subscribe(topic.GetTopicFromSubscription(subscription, topic.DataEvents))
	if err != nil {
		return err
	}

	newOutput.Start(output.Config(subscription.OutputConfig), sub.GetChan())
	manager.running[subscription.ID] = subscriberEntry{
		sub:             sub,
		output:          newOutput,
		geoSubscription: geoSubscription,
	}

	return nil
}

func (manager *natsManager) Stop(subscriptionID int64) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	v, exists := manager.running[subscriptionID]
	if !exists {
		return errors.New("Unknown subscription")
	}
	delete(manager.running, subscriptionID)
	unsubscribeSubscription(v.sub)
	v.output.Stop(stopTimeout)
	return nil
}

const stopTimeout = 3 * time.Second

func (manager *natsManager) Shutdown() {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for i, subscription := range manager.running {
		unsubscribeSubscription(subscription.sub)
		subscription.output.Stop(stopTimeout)
		delete(manager.running, i)
	}

	err := manager.publisher.Drain()
	log.Error("Failed to drain publisher", err)

	err = manager.publisher.Conn.Drain()
	if err != nil {
		log.Error("Failed to drain connection")
	}

	manager.natsServer.Shutdown()
}

func (manager *natsManager) Get(subscriptionID int64) (output.GeoSubscription, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	ret, exists := manager.running[subscriptionID]
	if !exists {
		return output.GeoSubscription{}, errors.New("unknown output")
	}
	return ret.geoSubscription, nil
}

func (manager *natsManager) Publish(topic topic.Topic, event event.PublishableEvent) {
	err := manager.publisher.Publish(topic.TopicString(), event)
	if err != nil {
		log.Error("Failed to publish message to NATS", err)
	}
}

func (manager *natsManager) Subscribe(topic topic.Topic) (Subscription, error) {
	return NewNatsSubscription(topic, manager.publisher)
}

// createNATSStreamingServer will attempt to start up a NATS streaming server
func createNATSStreamingServer(config NATSManagerConfig) (*stand.StanServer, error) {
	// NATS streaming server options
	natsOpts := stand.NewNATSOptions()
	natsOpts.Port = config.Port
	natsOpts.MaxPayload = 10000

	// NATS server options
	natsServerOptions := stand.GetDefaultOptions()
	natsServerOptions.ID = "output-manager"
	natsServerOptions.MaxSubscriptions = 1000000
	natsServerOptions.MaxMsgs = 10000000
	natsServerOptions.EnableLogging = config.Logging

	// NATS store directory (OPTIONAL)
	switch config.StoreType {
	case "file":
		natsServerOptions.StoreType = stores.TypeFile
		natsServerOptions.FileStoreOpts = stores.DefaultFileStoreOptions

		natsServerOptions.FilestoreDir = config.FileStoreDir
	default:
		natsServerOptions.StoreType = stores.TypeMemory
	}

	return stand.RunServerWithOpts(natsServerOptions, natsOpts)
}

func unsubscribeSubscription(subscription Subscription) {
	err := subscription.Unsubscribe()
	if err != nil {
		log.WithError(err).Errorf("Failed to unsubscribe subscription %v", subscription)
	}
}
