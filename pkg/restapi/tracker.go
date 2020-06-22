package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/gorilla/mux"
)

func (s *Server) createTracker(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	tracker, err := validation.GetTrackerFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	tracker.CollectionID = collectionID

	newTrackerID, err := validation.CreateTracker(tracker.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newTracker, err := validation.GetTracker(newTrackerID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newTracker.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Tracker, newTracker.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.CreatedEvent, event.TrackerEntity, newTracker.ID),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateTracker(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	trackerID, err := validation.GetTrackerID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	trackerBody, err := validation.GetTrackerFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// We do this to ensure that the ID of the original tracker is the only one who's being changed
	trackerBody.CollectionID = collectionID
	trackerBody.ID = trackerID

	err = validation.UpdateTracker(trackerBody.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updatedTracker, err := validation.GetTracker(trackerID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := updatedTracker.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Tracker, updatedTracker.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.TrackerEntity, updatedTracker.ID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getTracker(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	tracker, err := validation.GetTrackerFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := tracker.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteTracker(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	trackerID, err := validation.GetTrackerID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = validation.DeleteTracker(trackerID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Tracker, trackerID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.DeletedEvent, event.TrackerEntity, trackerID),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listTrackers(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	trackers, err := validation.ListTrackersByCollectionID(collectionID, userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(trackers)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) createTrackerPos(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	trackerID, err := validation.GetTrackerID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	position, err := validation.GetPositionFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	position.TrackerID = trackerID

	newPositionID, err := validation.CreatePosition(position.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newPosition, err := validation.GetPosition(newPositionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newPosition.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	// Publish collection and tracker data
	s.manager.Publish(
		topic.NewEntityTopic(topic.Collection, collectionID, topic.DataEvents),
		event.NewPositionEvent(collectionID, *newPosition.ToModel()),
	)
	s.manager.Publish(
		topic.NewEntityTopic(topic.Tracker, trackerID, topic.DataEvents),
		event.NewPositionEvent(collectionID, *newPosition.ToModel()),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getTrackerPosition(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	position, err := validation.GetPositionFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := position.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteTrackerPosition(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	positionID, err := validation.GetPositionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = validation.DeletePosition(positionID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listTrackerPositions(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	trackerID, err := validation.GetTrackerID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	positions, err := validation.ListPositionsByTrackerID(trackerID, userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(positions)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
