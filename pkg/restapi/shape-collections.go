package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/gorilla/mux"
)

func (s *Server) createShapeCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionBody, err := validation.GetShapeCollectionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newShapeCollectionID, err := validation.CreateShapeCollection(shapeCollectionBody.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newShapeCollection, err := validation.GetShapeCollection(newShapeCollectionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newShapeCollection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, newShapeCollection.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.CreatedEvent, event.ShapeCollectionEntity, newShapeCollection.ID),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateShapeCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeCollectionBody, err := validation.GetShapeCollectionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// We do this to ensure that the ID of the original shape collection is the only one who's being changed
	shapeCollectionBody.ID = shapeCollectionID

	err = validation.UpdateShapeCollection(shapeCollectionBody.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updatedShapeCollection, err := validation.GetShapeCollection(shapeCollectionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := updatedShapeCollection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, updatedShapeCollection.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.ShapeCollectionEntity, updatedShapeCollection.ID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getShapeCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollection, err := validation.GetShapeCollectionFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := shapeCollection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteShapeCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = s.store.DeleteShapeCollection(shapeCollectionID, userProfile.ID)
	if err != nil {
		handleError(err, w, log)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, shapeCollectionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.DeletedEvent, event.ShapeCollectionEntity, shapeCollectionID),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listShapeCollections(w http.ResponseWriter, r *http.Request) {
	userProfile := s.UserFromRequest(r)
	log := s.RequestLogger(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeCollections, err := validation.ListShapeCollections(userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(shapeCollections)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
