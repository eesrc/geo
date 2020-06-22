package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/store/errors"
)

// GetTrackerID returns a trackerID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetTrackerID(handlerParams HandlerParameterMap) (int64, error) {
	trackerID, err := handlerParams.TrackerID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("trackerId", fmt.Sprintf("The tracker id '%s' is malformed", handlerParams["trackerID"])),
		))
	}

	return trackerID, nil
}

// GetTrackerFromHandlerParams validates tracker (if) found in params and returns a tracker from store.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetTrackerFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*service.Tracker, error) {
	trackerID, err := GetTrackerID(handlerParams)
	if err != nil {
		return &service.Tracker{}, err
	}

	tracker, err := GetTracker(trackerID, userID, store)
	if err != nil {
		return &service.Tracker{}, err
	}

	return tracker, nil
}

// GetTracker returns a tracker from store. Returns a validation error or regular error
// if the fetch fails.
func GetTracker(trackerID int64, userId int64, store store.Store) (*service.Tracker, error) {
	tracker, err := store.GetTrackerByUserID(trackerID, userId)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return &service.Tracker{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("trackerId", fmt.Sprintf("The tracker with id '%d' might not exist", trackerID)),
				))
			}
		}

		return &service.Tracker{}, err
	}

	return service.NewTrackerFromModel(tracker), nil
}

// CreateTracker tries to create a tracker and returns a validation error or regular error
// if the creation fails
func CreateTracker(tracker *model.Tracker, userID int64, store store.Store) (int64, error) {
	newTrackerID, err := store.CreateTracker(tracker, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return -1, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("tracker.collectionId", fmt.Sprintf("The collection with id '%d' might not exist", tracker.CollectionID)),
				))
			}
		}
	}

	return newTrackerID, err
}

// UpdateTracker tries to update a tracker and returns a validation error or regular error
// if the update fails
func UpdateTracker(tracker *model.Tracker, userID int64, store store.Store) error {
	err := store.UpdateTracker(tracker, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("tracker.id", fmt.Sprintf("The tracker id '%d' might not exist", tracker.ID)),
					NewParameterErrorDetail("tracker.collectionId", fmt.Sprintf("The collection with id '%d' might not exist", tracker.CollectionID)),
				))
			}
		}
	}

	return err
}

// DeleteTracker tries to delete a tracker and returns a validation error or regular error
// if the delete fails
func DeleteTracker(trackerID int64, userID int64, store store.Store) error {
	err := store.DeleteTracker(trackerID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("trackerId", fmt.Sprintf("The tracker with id '%d' might not exist", trackerID)),
					NewParameterErrorDetail("collectionId", fmt.Sprintf("The collection with id '%d' might not exist", trackerID)),
				))
			}
		}
	}

	return err
}

// ListCollections lists collections based on userID and returns a store error if the list fails
func ListTrackersByCollectionID(collectionID int64, userID int64, filterParams FilterParams, store store.Store) ([]*service.Tracker, error) {
	trackers, err := store.ListTrackersByCollectionID(collectionID, userID, filterParams.Offset, filterParams.Limit)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return []*service.Tracker{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("collectionId", fmt.Sprintf("The collection with id '%d' might not exist", collectionID)),
				))
			}
		}

		return []*service.Tracker{}, err
	}

	var trackerList []*service.Tracker = make([]*service.Tracker, len(trackers))

	for i, tracker := range trackers {
		trackerList[i] = service.NewTrackerFromModel(&tracker)
	}

	return trackerList, nil
}

// GetTrackerFromBody retrieves a tracker from given body and decodes it.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetTrackerFromBody(body io.ReadCloser) (*service.Tracker, error) {
	jsonDecoder := json.NewDecoder(body)
	var tracker service.Tracker
	err := jsonDecoder.Decode(&tracker)
	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &tracker, getUnmarshalError(err)
		}

		return &tracker, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("tracker", "You need to provide a valid tracker object"),
			),
		)
	}

	return &tracker, nil
}
