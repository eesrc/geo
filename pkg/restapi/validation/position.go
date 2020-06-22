package validation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/store/errors"
)

// GetPositionID returns a positionID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetPositionID(handlerParams HandlerParameterMap) (int64, error) {
	positionID, err := handlerParams.PositionID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("positionId", fmt.Sprintf("The position id '%s' is malformed", handlerParams["trackerID"])),
		))
	}

	return positionID, nil
}

// GetPositionFromHandlerParams validates position (if) found in params and returns a position from store.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetPositionFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*service.Position, error) {
	positionID, err := GetPositionID(handlerParams)
	if err != nil {
		return &service.Position{}, err
	}

	position, err := GetPosition(positionID, userID, store)
	if err != nil {
		return &service.Position{}, err
	}

	return position, nil
}

// GetPosition returns a position from store. Returns a validation error or regular error
// if the fetch fails.
func GetPosition(positionID int64, userId int64, store store.Store) (*service.Position, error) {
	position, err := store.GetPositionByUserID(positionID, userId)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return &service.Position{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("positionId", fmt.Sprintf("The position with id '%d' might not exist", positionID)),
				))
			}
		}

		return &service.Position{}, err
	}

	return service.NewPositionFromModel(position), nil
}

// CreatePosition tries to create a position and returns a validation error or regular error
// if the creation fails
func CreatePosition(position *model.Position, userID int64, store store.Store) (int64, error) {
	newPositionID, err := store.CreatePosition(position, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return -1, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("position.trackerId", fmt.Sprintf("The tracker with id '%d' might not exist", position.TrackerID))))
			}
		}
	}

	return newPositionID, err
}

// DeletePosition tries to delete a position and returns a validation error or regular error
// if the delete fails
func DeletePosition(positionID int64, userID int64, store store.Store) error {
	err := store.DeletePosition(positionID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("positionId", fmt.Sprintf("The position with id '%d' might not exist", positionID)),
				))
			}
		}
	}

	return err
}

// ListPositionsByTrackerID lists collections based on userID and returns a store error if the list fails
func ListPositionsByTrackerID(trackerID int64, userID int64, filterParams FilterParams, store store.Store) ([]*service.Position, error) {
	positions, err := store.ListPositionsByTrackerID(trackerID, userID, filterParams.Offset, filterParams.Limit)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return []*service.Position{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("trackerId", fmt.Sprintf("The tracker id '%d' might not exist", trackerID)),
				))
			}
		}

		return []*service.Position{}, err
	}

	var positionList []*service.Position = make([]*service.Position, len(positions))

	for i, position := range positions {
		positionList[i] = service.NewPositionFromModel(&position)
	}

	return positionList, nil
}

// GetPositionFromBody retrieves a position from given body and decodes it.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetPositionFromBody(body io.ReadCloser) (*service.Position, error) {
	jsonDecoder := json.NewDecoder(body)
	position := service.NewPosition()

	err := jsonDecoder.Decode(&position)
	if err != nil {
		switch err := err.(type) {
		case *json.UnmarshalTypeError:
			return &position, getUnmarshalError(err)
		case base64.CorruptInputError:
			return &position, newError(
				NewErrorResponse(
					http.StatusBadRequest,
					NewParameterErrorDetail("payload", "The payload provided needs to be base64 encoded"),
				),
			)
		default:
			return &position, newError(
				NewErrorResponse(
					http.StatusBadRequest,
					NewParameterErrorDetail("position", "You need to provide a valid position object"),
				),
			)
		}
	}

	var fieldErrors []ErrorDetail

	if position.Lat == nil {
		fieldErrors = append(fieldErrors, NewParameterErrorDetail("lat", "The latitude of the position must be set"))
	}

	if position.Long == nil {
		fieldErrors = append(fieldErrors, NewParameterErrorDetail("lng", "The longitude of the position must be set"))
	}

	if position.Alt == nil {
		alt := 0.0
		position.Alt = &alt
	}

	if position.Precision == nil {
		precision := 1.0
		position.Precision = &precision
	}

	if *position.Precision > float64(1) || *position.Precision < float64(0) {
		fieldErrors = append(fieldErrors, NewParameterErrorDetail("precision", "The precision must be between 0 and 1"))
	}

	if len(fieldErrors) > 0 {
		return &position, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				fieldErrors...,
			),
		)
	}

	return &position, nil
}
