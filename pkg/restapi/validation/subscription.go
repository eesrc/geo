package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/store/errors"
	"github.com/eesrc/geo/pkg/sub"
)

// GetSubscriptionID returns a SubscriptionID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetSubscriptionID(handlerParams HandlerParameterMap) (int64, error) {
	subscriptionID, err := handlerParams.SubscriptionID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("subscriptionId", fmt.Sprintf("The subscription id '%s' is malformed", handlerParams["subscriptionID"])),
		))
	}

	return subscriptionID, nil
}

// GetSubscriptionFromHandlerParams validates subscription (if) found in params and returns a subscription from store.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetSubscriptionFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*service.Subscription, error) {
	subscriptionID, err := GetSubscriptionID(handlerParams)
	if err != nil {
		return &service.Subscription{}, err
	}

	subscription, err := GetSubscription(subscriptionID, userID, store)
	if err != nil {
		return &service.Subscription{}, err
	}

	return subscription, nil
}

// GetSubscription returns a subscription from store. Returns a validation error or regular error
// if the fetch fails.
func GetSubscription(subscriptionID int64, userId int64, store store.Store) (*service.Subscription, error) {
	subscription, err := store.GetSubscriptionByUserID(subscriptionID, userId)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return &service.Subscription{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("subscription.id", fmt.Sprintf("The subscription with id '%d' might not exist", subscriptionID)),
				))
			}
		}

		return &service.Subscription{}, err
	}

	return service.NewSubscriptionFromModel(subscription), nil
}

// CreateSubscription tries to create a collection and returns a validation error or regular error
// if the creation fails
func CreateSubscription(subscription *service.Subscription, userID int64, store store.Store) (int64, error) {
	newSubscriptionID, err := store.CreateSubscription(subscription.ToModel(), userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return -1, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("subscription.teamId", fmt.Sprintf("The provided teamId '%d' might not exist", *subscription.TeamID)),
					NewParameterErrorDetail("subscription.shapeCollectionId", fmt.Sprintf("The provided shapeCollectionId '%d' might not exist", *subscription.ShapeCollectionID)),
					getTrackableNotFoundErrorDetail(subscription),
				))
			}
		}
	}

	return newSubscriptionID, err
}

// UpdateSubscription tries to update a subscription and returns a validation error or regular error
// if the update fails
func UpdateSubscription(subscription *service.Subscription, userID int64, store store.Store) error {
	err := store.UpdateSubscription(subscription.ToModel(), userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("subscription.teamId", fmt.Sprintf("The provided teamId '%d' might not exist", *subscription.TeamID)),
					NewParameterErrorDetail("subscription.shapeCollectionId", fmt.Sprintf("The provided shapeCollectionId '%d' might not exist", *subscription.ShapeCollectionID)),
					getTrackableNotFoundErrorDetail(subscription),
				))
			}
		}
	}

	return err
}

// DeleteSusbcription tries to delete a subscription and returns a validation error or regular error
// if the delete fails
func DeleteSubscription(subscriptionID int64, userID int64, store store.Store) error {
	err := store.DeleteSubscription(subscriptionID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("subscriptionId", fmt.Sprintf("The subscription with id '%d' might not exist", subscriptionID)),
				))
			}
		}
	}

	return err
}

// ListSubscriptions lists subscriptions based on userID and returns a store error if the list fails
func ListSubscriptions(userID int64, filterParams FilterParams, store store.Store) ([]*service.Subscription, error) {
	subscriptions, err := store.ListSubscriptionsByUserID(userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.Subscription{}, err
	}

	var subscriptionList []*service.Subscription = make([]*service.Subscription, len(subscriptions))

	for i, susbcription := range subscriptions {
		subscriptionList[i] = service.NewSubscriptionFromModel(&susbcription)
	}

	return subscriptionList, nil
}

// ListSubscriptionsByTrackerID lists subscriptions based on userID and trackerID and returns a store error if the list fails
func ListSubscriptionsByTrackerID(trackerID int64, userID int64, filterParams FilterParams, store store.Store) ([]*service.Subscription, error) {
	subscriptions, err := store.ListSubscriptionsByTrackerID(trackerID, userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.Subscription{}, err
	}

	var subscriptionList []*service.Subscription = make([]*service.Subscription, len(subscriptions))

	for i, susbcription := range subscriptions {
		subscriptionList[i] = service.NewSubscriptionFromModel(&susbcription)
	}

	return subscriptionList, nil
}

// ListSubscriptionsByCollectionID lists subscriptions based on userID and trackerID and returns a store error if the list fails
func ListSubscriptionsByCollectionID(collectionID int64, userID int64, filterParams FilterParams, store store.Store) ([]*service.Subscription, error) {
	subscriptions, err := store.ListSubscriptionsByCollectionID(collectionID, userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.Subscription{}, err
	}

	var subscriptionList []*service.Subscription = make([]*service.Subscription, len(subscriptions))

	for i, susbcription := range subscriptions {
		subscriptionList[i] = service.NewSubscriptionFromModel(&susbcription)
	}

	return subscriptionList, nil
}

// GetSubscriptionFromBody retrieves a subscription from given body and decodes it.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetSubscriptionFromBody(body io.ReadCloser) (*service.Subscription, error) {
	jsonDecoder := json.NewDecoder(body)

	subscription := service.NewSubscription()

	err := jsonDecoder.Decode(&subscription)
	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &subscription, getUnmarshalError(err)
		}

		return &subscription, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("subscription", "You need to provide a valid subscription object"),
			),
		)
	}

	var fieldErrors []ErrorDetail

	if !isValidOutputType(subscription.Output.Type) {
		validOutputs := []string{}
		for _, output := range sub.ValidOutputTypes {
			validOutputs = append(validOutputs, "'"+string(output)+"'")
		}

		fieldErrors = append(fieldErrors, NewParameterErrorDetail(
			"subscription.output",
			fmt.Sprintf("The output '%s' is not valid. Available outputs are: %s", subscription.Output, strings.Join(validOutputs, ", "))),
		)
	}

	if !isValidTrackableType(subscription.Trackable.Type) {
		validTrackableTypes := []string{}
		for _, validType := range sub.ValidTrackableTypes {
			validTrackableTypes = append(validTrackableTypes, "'"+string(validType)+"'")
		}

		fieldErrors = append(fieldErrors, NewParameterErrorDetail(
			"subscription.trackable.type",
			fmt.Sprintf("The trackableType '%s' is not valid. Available types are: %s", subscription.Trackable.Type, strings.Join(validTrackableTypes, ", "))),
		)
	}

	if !isValidMovementSubscriptions(subscription.TriggerCriteria.TriggerTypes) {

		validTriggerCriterias := []string{}
		for _, validMovementType := range sub.ValidMovementTypes {
			validTriggerCriterias = append(validTriggerCriterias, "'"+string(validMovementType)+"'")
		}

		fieldErrors = append(fieldErrors, NewParameterErrorDetail(
			"subscription.types",
			fmt.Sprintf(
				"The list of trigger types '%s' is not valid. Available types are: %s",
				subscription.TriggerCriteria.TriggerTypes, strings.Join(validTriggerCriterias, ", "),
			)),
		)
	}

	if subscription.TeamID == nil {
		fieldErrors = append(fieldErrors, NewParameterErrorDetail("subscription.teamId", fmt.Sprintf("You need to provide a teamId")))
	}

	if subscription.ShapeCollectionID == nil {
		fieldErrors = append(fieldErrors, NewParameterErrorDetail("subscription.shapeCollectionId", fmt.Sprintf("You need to provide a shapeCollectionId")))
	}

	if subscription.Trackable.ID == nil {
		fieldErrors = append(fieldErrors, NewParameterErrorDetail("subscription.trackable.id", fmt.Sprintf("You need to provide a trackableId")))
	}

	if len(fieldErrors) > 0 {
		return &subscription, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				fieldErrors...,
			),
		)
	}

	return &subscription, nil
}

func isValidOutputType(outputType sub.OutputType) bool {
	for _, validType := range sub.ValidOutputTypes {
		if validType == outputType {
			return true
		}
	}

	return false
}

func isValidTrackableType(trackableType sub.TrackableType) bool {
	for _, validTrackableType := range sub.ValidTrackableTypes {
		if validTrackableType == trackableType {
			return true
		}
	}
	return false
}

func isValidMovementSubscriptions(movements sub.MovementList) bool {
	for _, movement := range movements {
		valid := false

		for _, validMovementSub := range sub.ValidMovementTypes {
			if movement == validMovementSub {
				valid = true
				continue
			}
		}

		if !valid {
			return false
		}
	}

	return true
}

func getTrackableNotFoundErrorDetail(subscription *service.Subscription) ErrorDetail {
	switch subscription.Trackable.Type {
	case sub.Collection:
		return NewParameterErrorDetail("subscription.trackable.id", fmt.Sprintf("The provided trackableId (collection) '%d' might not exist", *subscription.Trackable.ID))
	case "tracker":
		return NewParameterErrorDetail("subscription.trackable.id", fmt.Sprintf("The provided trackableId (tracker) '%d' might not exist", *subscription.Trackable.ID))
	default:
		return NewParameterErrorDetail("subscription.trackable.id", fmt.Sprintf("The provided trackableId '%d' might not exist", *subscription.Trackable.ID))
	}
}
