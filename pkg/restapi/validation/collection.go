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

// GetCollectionID returns a collectionID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetCollectionID(handlerParams HandlerParameterMap) (int64, error) {
	collectionID, err := handlerParams.CollectionID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("collectionId", fmt.Sprintf("The collection id '%s' is malformed", handlerParams["collectionID"])),
		))
	}

	return collectionID, nil
}

// GetCollectionFromHandlerParams validates collection ID (if) found in params and returns a collection from store.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetCollectionFromHandlerParams(handlerParams HandlerParameterMap, userId int64, store store.Store) (*service.Collection, error) {
	collectionID, err := GetCollectionID(handlerParams)
	if err != nil {
		return &service.Collection{}, err
	}

	collection, err := GetCollection(collectionID, userId, store)

	if err != nil {
		return &service.Collection{}, err
	}

	return collection, nil
}

// GetCollection returns a collection from store. Returns a validation error or regular error
// if the fetch fails.
func GetCollection(collectionID int64, userId int64, store store.Store) (*service.Collection, error) {
	collection, err := store.GetCollectionByUserID(collectionID, userId)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return &service.Collection{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("collectionId", fmt.Sprintf("The collection with id '%d' might not exist", collectionID)),
				))
			}
		}

		return &service.Collection{}, err
	}

	return service.NewCollectionFromModel(collection), nil
}

// CreateCollection tries to create a collection and returns a validation error or regular error
// if the creation fails
func CreateCollection(collection *model.Collection, userID int64, store store.Store) (int64, error) {
	newCollectionID, err := store.CreateCollection(collection, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return -1, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("collection.teamId", fmt.Sprintf("The team with id '%d' might not exist", collection.TeamID)),
				))
			}
		}
	}

	return newCollectionID, err
}

// UpdateCollection tries to update a collection and returns a validation error or regular error
// if the update fails
func UpdateCollection(collection *model.Collection, userID int64, store store.Store) error {
	err := store.UpdateCollection(collection, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("collection.id", fmt.Sprintf("The collection id '%d' might not exist", collection.ID)),
					NewParameterErrorDetail("collection.teamId", fmt.Sprintf("The team with id '%d' might not exist", collection.TeamID)),
				))
			}
		}
	}

	return err
}

// DeleteCollection tries to delete a collection and returns a validation error or regular error
// if the delete fails
func DeleteCollection(collectionID int64, userID int64, store store.Store) error {
	err := store.DeleteCollection(collectionID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("collectionId", fmt.Sprintf("The collection with id '%d' might not exist", collectionID)),
				))
			}
		}
	}

	return err
}

// ListCollections lists collections based on userID and returns a store error if the list fails
func ListCollections(userID int64, filterParams FilterParams, store store.Store) ([]*service.Collection, error) {
	collections, err := store.ListCollectionsByUserID(userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.Collection{}, err
	}

	var collectionList []*service.Collection = make([]*service.Collection, len(collections))

	for i, collection := range collections {
		collectionList[i] = service.NewCollectionFromModel(&collection)
	}

	return collectionList, nil
}

// GetCollectionFromBody retrieves a collection from given body and decodes it.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetCollectionFromBody(body io.ReadCloser) (*service.Collection, error) {
	jsonDecoder := json.NewDecoder(body)
	var collection service.Collection
	err := jsonDecoder.Decode(&collection)
	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &collection, getUnmarshalError(err)
		}

		return &collection, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("collection", "You need to provide a valid collection object"),
			),
		)
	}

	return &collection, nil
}
