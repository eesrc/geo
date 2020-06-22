package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/gorilla/mux"
)

func (s *Server) getFeatureCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	featureCollection, err := validation.GetShapesAsFeatureCollection(shapeCollectionID, userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := featureCollection.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateFeatureCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Retrieve and triangulate shapes from body
	newShapes, err := validation.GetShapesFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Create models from shapes
	shapeModels := make([]*model.Shape, len(newShapes))
	for i, newShape := range newShapes {
		shapeModels[i] = &model.Shape{
			Name:              newShape.GetName(),
			ShapeCollectionID: shapeCollectionID,
			Properties:        newShape.GetProperties(),
			Shape:             newShape,
		}
	}

	// Fully replace the old shapes with new shapes
	err = validation.ReplaceShapesInShapeCollection(shapeCollectionID, userProfile.ID, shapeModels, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Retrieve the newly updated shapes
	updateFeatureCollection, err := validation.GetShapesAsFeatureCollection(shapeCollectionID, userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updateShapeCollectionGeoSubscriptions(shapeCollectionID, s.manager, s.store)

	jsonBytes, err := json.Marshal(updateFeatureCollection)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, shapeCollectionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.ShapeCollectionEntity, shapeCollectionID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) createShape(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeBody, err := validation.GetShapeFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeModel := &model.Shape{
		Name:              shapeBody.GetName(),
		ShapeCollectionID: shapeCollectionID,
		Shape:             shapeBody,
	}

	newShapeID, err := s.store.CreateShape(shapeModel, userProfile.ID)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updateShapeCollectionGeoSubscriptions(shapeCollectionID, s.manager, s.store)

	newShape, err := validation.GetShape(shapeCollectionID, newShapeID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newShape.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, shapeCollectionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.ShapeCollectionEntity, shapeCollectionID),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateShapeGeoJSON(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeID, err := validation.GetShapeID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeBody, err := validation.GetShapeFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Create model from shape
	shapeModel := &model.Shape{
		ID:                shapeID,
		Name:              shapeBody.GetName(),
		ShapeCollectionID: shapeCollectionID,
	}

	err = validation.UpdateShape(shapeModel, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updateShapeCollectionGeoSubscriptions(shapeCollectionID, s.manager, s.store)

	updatedShape, err := validation.GetShape(shapeCollectionID, shapeID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := updatedShape.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, shapeCollectionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.ShapeCollectionEntity, shapeCollectionID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getShape(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shape, err := validation.GetShapeFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := shape.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getShapeGeoJSON(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shape, err := validation.GetShapeGeoJSONFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := shape.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteShape(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	shapeCollectionID, err := validation.GetShapeCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeID, err := validation.GetShapeID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = s.store.DeleteShape(shapeCollectionID, shapeID, userProfile.ID)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updateShapeCollectionGeoSubscriptions(shapeCollectionID, s.manager, s.store)

	s.manager.Publish(
		topic.NewEntityTopic(topic.ShapeCollections, shapeCollectionID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.ShapeCollectionEntity, shapeCollectionID),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listShapesByCollection(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapeCollection, err := validation.GetShapeCollectionFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	shapes, err := validation.ListShapes(shapeCollection.ID, userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(shapes)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
