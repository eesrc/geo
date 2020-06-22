package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/gorilla/mux"
)

func (s *Server) createCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionBody, err := validation.GetCollectionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newCollectionID, err := validation.CreateCollection(collectionBody.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newCollection, err := validation.GetCollection(newCollectionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newCollection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Collection, newCollection.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.CreatedEvent, event.CollectionEntity, newCollection.ID),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	collectionBody, err := validation.GetCollectionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Set the collectionID based on path
	collectionBody.ID = collectionID

	err = validation.UpdateCollection(collectionBody.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updatedCollection, err := validation.GetCollection(collectionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := updatedCollection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Collection, updatedCollection.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.CollectionEntity, updatedCollection.ID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collection, err := validation.GetCollectionFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := collection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = validation.DeleteCollection(collectionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Collection, collectionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.DeletedEvent, event.CollectionEntity, collectionID),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listCollections(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	userProfile := s.UserFromRequest(r)

	collectionList, err := validation.ListCollections(userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(collectionList)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
