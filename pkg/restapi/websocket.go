package restapi

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/sub/manager"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const handshakeTimeout = 60 * time.Second
const keepAliveTimeout = 10 * time.Second

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	HandshakeTimeout:  handshakeTimeout,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) trackerWebsocketData(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	// As a simple check that the user has access to the tracker, we retrieve it with the user ID
	tracker, err := validation.GetTrackerFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Upgrade HTTP request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Error upgrading web socket")
		return
	}

	// Create subscription of data events on tracker
	topic := topic.NewEntityTopic(topic.Tracker, tracker.ID, topic.DataEvents)
	subscription, err := s.manager.Subscribe(topic)
	defer unsubscribeSubscription(subscription)

	if err != nil {
		log.WithError(err).Errorf("Could not get a subscription on given topic '%s'", topic.TopicString())
		return
	}

	initiateWebsocketSubscription(conn, subscription.GetChan(), genericMessageHandler)
}

func (s *Server) collectionWebsocketData(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	var handlerParams validation.HandlerParameterMap = mux.Vars(r)

	collection, err := validation.GetCollectionFromHandlerParams(handlerParams, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Upgrade HTTP request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Error upgrading web socket")
		return
	}

	// Create subscription of events on collection
	topic := topic.NewEntityTopic(topic.Collection, collection.ID, topic.AllEvents)
	subscription, err := s.manager.Subscribe(topic)
	defer unsubscribeSubscription(subscription)

	if err != nil {
		log.WithError(err).Errorf("Could not get a subscription on given topic '%s'", topic.TopicString())
		return
	}

	initiateWebsocketSubscription(conn, subscription.GetChan(), genericMessageHandler)
}

func (s *Server) subscriptionWebsocketData(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	subscription, err := validation.GetSubscriptionFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Upgrade HTTP request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Error upgrading web socket")
		return
	}

	// Create subscription of events on subscription
	topic := topic.NewEntityTopic(topic.Subscription, subscription.ID, topic.AllEvents)
	topicSubscription, err := s.manager.Subscribe(topic)
	defer unsubscribeSubscription(topicSubscription)

	if err != nil {
		log.WithError(err).Errorf("Could not get a subscription on given topic '%s'", topic.TopicString())
		return
	}

	initiateWebsocketSubscription(conn, topicSubscription.GetChan(), genericMessageHandler)
}

func genericMessageHandler(message interface{}) (interface{}, error) {
	genericEvent, err := event.DecodeEvent(message)
	if err != nil {
		return genericEvent, err
	}

	switch decodedEvent := genericEvent.(type) {
	case event.LifeCycleEvent:
		return service.NewLifecycleEventFromModel(&decodedEvent), nil
	case event.PositionEvent:
		return service.NewPositionEventFromModel(&decodedEvent), nil
	case event.SubscriptionEvent:
		return service.NewSubscriptionEventFromModel(&decodedEvent), nil
	default:
		return genericEvent, fmt.Errorf("Could not handle the message of type %T. %v", decodedEvent, decodedEvent)
	}
}

func initiateWebsocketSubscription(websocketConnection *websocket.Conn, channel <-chan interface{}, handler func(interface{}) (interface{}, error)) {
	// Add websocket closing channel
	defer websocketConnection.Close()
	close := make(chan bool)

	// Listen to websocket close event
	go func(c *websocket.Conn) {
		for {
			if _, _, err := c.NextReader(); err != nil {
				log.WithError(err).Info("Closing websocket")
				c.Close()
				close <- true
				return
			}
		}
	}(websocketConnection)

	// Set up data subscription
	for {
		select {
		case <-close:
			// Close event from websocket, return
			return
		case msg, open := <-channel:
			// Got a message
			if !open {
				return
			}

			payload, err := handler(msg)
			if err != nil {
				log.Error("Error when handling payload", err)
				continue
			}

			err = websocketConnection.SetWriteDeadline(time.Now().Add(keepAliveTimeout))
			if err != nil {
				return
			}

			if err := websocketConnection.WriteJSON(payload); err != nil {
				return
			}
		case <-time.After(keepAliveTimeout):
			// send KeepAlive-messages
			err := websocketConnection.SetWriteDeadline(time.Now().Add(keepAliveTimeout))
			if err != nil {
				return
			}

			if err := websocketConnection.WriteJSON(service.NewWebsocketKeepAlive()); err != nil {
				// Implicitly close connection when error on write
				return
			}
		}
	}
}

func unsubscribeSubscription(subscription manager.Subscription) {
	err := subscription.Unsubscribe()

	if err != nil {
		log.
			WithField("subscription", subscription).
			WithError(err).
			Error("Failed to unsubscribe subscription during websocket disconnect.")
	}
}
