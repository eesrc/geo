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

// GetShapeCollectionID returns a ShapeCollectionID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetShapeCollectionID(handlerParams HandlerParameterMap) (int64, error) {
	shapeCollectionID, err := handlerParams.ShapeCollectionID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("shapeCollectionId", fmt.Sprintf("The shape collection id '%s' is malformed", handlerParams["shapeCollectionID"])),
		))
	}

	return shapeCollectionID, nil
}

// GetShapeCollectionFromHandlerParams validates shape collection (if) found in params and returns a shape collection from store.
// Returns a validation error containing an ErrorResponse based on what went wrong
func GetShapeCollectionFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*service.ShapeCollection, error) {
	shapeCollectionID, err := GetShapeCollectionID(handlerParams)
	if err != nil {
		return &service.ShapeCollection{}, err
	}

	shapeCollection, err := GetShapeCollection(shapeCollectionID, userID, store)
	if err != nil {
		return &service.ShapeCollection{}, err
	}

	return shapeCollection, err
}

// GetShapeCollection returns a shape collection from store. Returns a validation error or regular error
// if the fetch fails.
func GetShapeCollection(shapeCollectionID int64, userID int64, store store.Store) (*service.ShapeCollection, error) {
	shapeCollection, err := store.GetShapeCollectionByUserID(shapeCollectionID, userID)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return &service.ShapeCollection{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shapeCollectionId", fmt.Sprintf("The shape collection with id '%d' might not exist", shapeCollectionID)),
				))
			}
		}
	}

	return service.NewShapeCollectionFromModel(shapeCollection), err
}

// CreateShapeCollection tries to create a shape collection and returns a validation error or regular error
// if the creation fails
func CreateShapeCollection(shapeCollection *model.ShapeCollection, userID int64, store store.Store) (int64, error) {
	newShapeCollectionID, err := store.CreateShapeCollection(shapeCollection, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return -1, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shapeCollection.teamId", fmt.Sprintf("The team with id '%d' might not exist", shapeCollection.TeamID)),
				))
			}
		}
	}

	return newShapeCollectionID, err
}

// UpdateShapeCollection tries to update a shape collection and returns a validation error or regular error
// if the update fails
func UpdateShapeCollection(shapeCollection *model.ShapeCollection, userID int64, store store.Store) error {
	err := store.UpdateShapeCollection(shapeCollection, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shapeCollection.id", fmt.Sprintf("The shape collection id '%d' might not exist", shapeCollection.ID)),
					NewParameterErrorDetail("shapeCollection.teamId", fmt.Sprintf("The team with id '%d' might not exist", shapeCollection.TeamID)),
				))
			}
		}
	}

	return err
}

// DeleteShapeCollection tries to delete a collection and returns a validation error or regular error
// if the delete fails
func DeleteShapeCollection(shapeCollectionID int64, userID int64, store store.Store) error {
	err := store.DeleteShapeCollection(shapeCollectionID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shapeCollection.id", fmt.Sprintf("The shape collection with id '%d' might not exist", shapeCollectionID)),
				))
			}
		}
	}

	return err
}

// ListShapeCollections lists shape collections based on userID and returns a store error if the list fails
func ListShapeCollections(userID int64, filterParams FilterParams, store store.Store) ([]*service.ShapeCollection, error) {
	shapeCollections, err := store.ListShapeCollectionsByUserID(userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.ShapeCollection{}, err
	}

	var shapeCollectionList []*service.ShapeCollection = make([]*service.ShapeCollection, len(shapeCollections))

	for i, shapeCollection := range shapeCollections {
		shapeCollectionList[i] = service.NewShapeCollectionFromModel(&shapeCollection)
	}

	return shapeCollectionList, nil
}

// GetShapeCollectionFromBody retrieves shape-collection from given body. Returns a validation error or regular error
// if the fetch fails.
func GetShapeCollectionFromBody(body io.ReadCloser) (*service.ShapeCollection, error) {
	jsonDecoder := json.NewDecoder(body)
	var shapeCollection service.ShapeCollection
	err := jsonDecoder.Decode(&shapeCollection)
	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &shapeCollection, getUnmarshalError(err)
		}

		return &shapeCollection, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("shapeCollection", "You need to provide a valid shapeCollection object"),
			),
		)
	}

	return &shapeCollection, nil
}
