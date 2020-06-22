package validation

import (
	"fmt"
	"net/http"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/store/errors"
	"github.com/eesrc/geo/pkg/tria/geometry"
	"github.com/eesrc/geo/pkg/tria/gj"
	geojson "github.com/paulmach/go.geojson"
)

// GetShapeID returns a shapeID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetShapeID(handlerParams HandlerParameterMap) (int64, error) {
	collectionID, err := handlerParams.ShapeID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("shapeId", fmt.Sprintf("The collection id '%s' is malformed", handlerParams["shapeID"])),
		))
	}

	return collectionID, nil
}

// GetShapeFromHandlerParams validates shape (if) found in params and returns a shape from store.
// Returns a validation error containing an ErrorResponse based on what went wrong
func GetShapeFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*service.Shape, error) {
	shapeCollectionID, err := GetShapeCollectionID(handlerParams)
	if err != nil {
		return &service.Shape{}, err
	}

	shapeID, err := handlerParams.ShapeID()
	if err != nil {
		return &service.Shape{}, err
	}

	shape, err := GetShape(shapeCollectionID, shapeID, userID, store)
	if err != nil {
		return &service.Shape{}, err
	}

	return shape, nil
}

// GetShape returns a shape from store. Returns a validation error or regular error
// if the fetch fails.
func GetShape(shapeCollectionID int64, shapeID int64, userID int64, store store.Store) (*service.Shape, error) {
	shape, err := store.GetShapeByUserID(shapeCollectionID, shapeID, userID, false)
	if err != nil {
		return &service.Shape{}, newError(NewErrorResponse(
			http.StatusNotFound,
			NewParameterErrorDetail("shapeId", fmt.Sprintf("The shape with id '%d' might not exist", shapeID)),
			NewParameterErrorDetail("shapeCollectionId", fmt.Sprintf("The shape collection with id '%d' might not exist", shapeCollectionID)),
		))
	}

	return service.NewShapeFromModel(shape), nil
}

// GetShape returns a shape from store. Returns a validation error or regular error
// if the fetch fails.
func GetShapeGeoJSON(shapeCollectionID int64, shapeID int64, userID int64, store store.Store) (*geojson.Feature, error) {
	shape, err := store.GetShapeByUserID(shapeCollectionID, shapeID, userID, true)
	if err != nil {
		return &geojson.Feature{}, newError(NewErrorResponse(
			http.StatusNotFound,
			NewParameterErrorDetail("shapeId", fmt.Sprintf("The shape with id '%d' might not exist", shapeID)),
			NewParameterErrorDetail("shapeCollectionId", fmt.Sprintf("The shape collection with id '%d' might not exist", shapeCollectionID)),
		))
	}

	geoJSONFeature := gj.GetFeaturesFromShape(shape.Shape, false)
	if len(geoJSONFeature) == 0 {
		return &geojson.Feature{}, nil
	}

	return geoJSONFeature[0], nil
}

// GetShapeFromHandlerParams validates shape (if) found in params and returns a GeoJSON Feature belonging to the shape from store.
// Returns a validation error containing an ErrorResponse based on what went wrong
func GetShapeGeoJSONFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*geojson.Feature, error) {
	shapeCollectionID, err := GetShapeCollectionID(handlerParams)
	if err != nil {
		return &geojson.Feature{}, err
	}

	shapeID, err := handlerParams.ShapeID()
	if err != nil {
		return &geojson.Feature{}, err
	}

	geoJSONFeature, err := GetShapeGeoJSON(shapeCollectionID, shapeID, userID, store)
	if err != nil {
		return &geojson.Feature{}, err
	}

	return geoJSONFeature, nil
}

// ReplaceShapesInShapeCollection takes in a list and replaces the existing shapes in an existing shape collection. Returns
// a validation error if the update and replace fails
func ReplaceShapesInShapeCollection(shapeCollectionID int64, userID int64, shapes []*model.Shape, store store.Store) error {
	err := store.ReplaceShapesInShapeCollection(shapeCollectionID, userID, shapes)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shape.collectionId", fmt.Sprintf("The collection with id '%d' might not exist", shapeCollectionID)),
				))
			}
		}
	}

	return err
}

// UpdateShape tries to update a shape and returns a validation error or regular error
// if the update fails
func UpdateShape(shape *model.Shape, userID int64, store store.Store) error {
	err := store.UpdateShape(shape, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shape.id", fmt.Sprintf("The shape id '%d' might not exist", shape.ID)),
					NewParameterErrorDetail("shape.collectionId", fmt.Sprintf("The shape collection with id '%d' might not exist", shape.ShapeCollectionID)),
				))
			}
		}
	}

	return err
}

// GetShapesAsFeatureCollection returns a feature collection based on shapeCollectionID and userID and returns a store error if the list fails
func GetShapesAsFeatureCollection(shapeCollectionID int64, userID int64, filterParams FilterParams, store store.Store) (geojson.FeatureCollection, error) {
	shapeModels, err := store.ListShapesByShapeCollectionIDAndUserID(shapeCollectionID, userID, true, filterParams.Offset, filterParams.Limit)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return geojson.FeatureCollection{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shape.collectionId", fmt.Sprintf("The shape collection with id '%d' might not exist", shapeCollectionID)),
				))
			}
		}

		return geojson.FeatureCollection{}, err
	}

	shapes := make([]geometry.Shape, len(shapeModels))

	for i, shapeModel := range shapeModels {
		shapes[i] = shapeModel.Shape
	}

	featureCollection := gj.NewGeoJSONFeatureCollectionFromShapes(shapes, false)

	return featureCollection, nil
}

// ListShapes lists the shapes in a shape collection based on given filters. Returns a validation error if the list fails.
func ListShapes(shapeCollectionID int64, userID int64, filterParams FilterParams, store store.Store) ([]*service.Shape, error) {
	shapes, err := store.ListShapesByShapeCollectionIDAndUserID(shapeCollectionID, userID, false, filterParams.Offset, filterParams.Limit)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return []*service.Shape{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("shape.collectionId", fmt.Sprintf("The shape collection with id '%d' might not exist", shapeCollectionID)),
				))
			}
		}
	}

	shapeList := make([]*service.Shape, len(shapes))

	for i, shape := range shapes {
		shapeList[i] = service.NewShapeFromModel(&shape)
	}

	return shapeList, err
}
