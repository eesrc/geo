package restapi

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/sub/manager"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/eesrc/geo/pkg/sub/output"
	"github.com/eesrc/geo/pkg/tria/index"
	"github.com/gorilla/mux"
)

func (s *Server) createSubscription(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	subscriptionBody, err := validation.GetSubscriptionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newSubscriptionID, err := validation.CreateSubscription(subscriptionBody, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newSubscription, err := validation.GetSubscription(newSubscriptionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	if newSubscription.Active {
		// The subscription is active, so we initiate a geoSubscription
		modelShapes, err := s.store.ListShapesByShapeCollectionIDAndUserID(*subscriptionBody.ShapeCollectionID, userProfile.ID, true, 0, 10000)
		if err != nil {
			handleError(err, w, log)
			return
		}

		index := index.NewRTreeIndexFromModel(modelShapes)

		geoSubscription := output.NewGeoSubscription(
			*newSubscription.ToModel(),
			index,
			s.store,
		)

		err = s.manager.Update(geoSubscription)

		if err != nil {
			handleSubscriptionUpdateError(geoSubscription, err)
		}
	}

	jsonBytes, err := newSubscription.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Subscription, newSubscription.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.CreatedEvent, event.SubscriptionEntity, newSubscription.ID),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getSubscription(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	subscription, err := validation.GetSubscriptionFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := subscription.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateSubscription(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	subscriptionID, err := validation.GetSubscriptionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	subscriptionBody, err := validation.GetSubscriptionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	subscriptionBody.ID = subscriptionID

	err = validation.UpdateSubscription(subscriptionBody, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updatedSubscription, err := validation.GetSubscription(subscriptionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// The subscription is potentially running/not running, so we initiate an update of a geoSubscription
	shapes, err := s.store.ListShapesByShapeCollectionIDAndUserID(*updatedSubscription.ShapeCollectionID, userProfile.ID, true, 0, 1000000)
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeIndex := index.NewRTreeIndexFromModel(shapes)

	movements, err := s.store.ListMovementsBySubscriptionID(updatedSubscription.ID, 0, 100)
	if err != nil {
		handleError(err, w, log)
		return
	}

	geoSubscription := output.NewGeoSubscriptionWithMovements(
		*updatedSubscription.ToModel(),
		shapeIndex,
		s.store,
		movements,
	)

	err = s.manager.Update(geoSubscription)

	if err != nil {
		handleSubscriptionUpdateError(geoSubscription, err)
	}

	jsonBytes, err := updatedSubscription.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Subscription, updatedSubscription.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.SubscriptionEntity, updatedSubscription.ID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	subscriptionID, err := validation.GetSubscriptionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = s.manager.Stop(subscriptionID)
	if err != nil {
		handleSubscriptionStopError(subscriptionID, err)
	}

	err = validation.DeleteSubscription(subscriptionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Subscription, subscriptionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.DeletedEvent, event.SubscriptionEntity, subscriptionID),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	subscriptions, err := validation.ListSubscriptions(userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(subscriptions)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

// updateShapeCollectionGeoSubscriptions will update all subscriptions reliant on the shape collection. The update will be launched
// in its own go routine. This update might fail, but should not affect the normal program flow
func updateShapeCollectionGeoSubscriptions(shapeCollectionID int64, manager manager.Manager, store store.Store) {
	go func() {
		geoSubscriptionModels, err := store.ListGeoSubscriptionsByShapeCollectionID(shapeCollectionID, 0, 100000)
		if err != nil {
			log.WithError(err).Errorf("Failed to list geo subscriptions for ShapeCollectionID %d", shapeCollectionID)
		}

		for _, geoSubscriptionModel := range geoSubscriptionModels {
			geoSubscription := output.NewGeoSubscriptionFromModel(geoSubscriptionModel, store)
			err = manager.Update(geoSubscription)

			if err != nil {
				handleSubscriptionUpdateError(geoSubscription, err)
			}
		}
	}()
}

func handleSubscriptionUpdateError(geoSubscription output.GeoSubscription, err error) {
	log.WithError(err).Errorf("Failed to update subscription %v", geoSubscription)
}

func handleSubscriptionStopError(subscriptionID int64, err error) {
	log.WithError(err).Errorf("Failed to stop subscription %d", subscriptionID)
}
