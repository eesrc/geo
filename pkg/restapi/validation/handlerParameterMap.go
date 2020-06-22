package validation

import (
	"fmt"
	"strconv"
)

//HandlerParameterMap is a list of parameter
type HandlerParameterMap map[string]string

// KeyNotFoundError is a error returned when no parameter was found on
// the parameter map
type KeyNotFoundError struct {
	field string
}

func (parameterNotFoundError *KeyNotFoundError) Error() string {
	return fmt.Sprintf("Could not find parameter %s", parameterNotFoundError.field)
}

// TokenID retrieves the tokenID from the parameterMap
func (parameterMap *HandlerParameterMap) TokenID() (string, error) {
	return parameterMap.AsString("tokenID")
}

// CollectionID retrieves the collectionID from the parameterMap
func (parameterMap *HandlerParameterMap) CollectionID() (int64, error) {
	return parameterMap.AsInt64("collectionID")
}

// TrackerID retrieves the collectionID from the parameterMap
func (parameterMap *HandlerParameterMap) TrackerID() (int64, error) {
	return parameterMap.AsInt64("trackerID")
}

// TeamID retrieves the TeamID from the parameterMap
func (parameterMap *HandlerParameterMap) TeamID() (int64, error) {
	return parameterMap.AsInt64("teamID")
}

// PositionID retrieves the PositionID from the parameterMap
func (parameterMap *HandlerParameterMap) PositionID() (int64, error) {
	return parameterMap.AsInt64("positionID")
}

// ShapeCollectionID retrieves the ShapeCollectionID from the parameterMap
func (parameterMap *HandlerParameterMap) ShapeCollectionID() (int64, error) {
	return parameterMap.AsInt64("shapeCollectionID")
}

// ShapeID retrieves the ShapeID from the parameterMap
func (parameterMap *HandlerParameterMap) ShapeID() (int64, error) {
	return parameterMap.AsInt64("shapeID")
}

// SubscriptionID retrieves the SubscriptionID from the parameterMap
func (parameterMap *HandlerParameterMap) SubscriptionID() (int64, error) {
	return parameterMap.AsInt64("subscriptionID")
}

// AsInt64 returns the parameter as an int64
func (parameterMap *HandlerParameterMap) AsInt64(id string) (int64, error) {
	v, ok := (*parameterMap)[id]

	if !ok {
		return -1, &KeyNotFoundError{field: id}
	}

	return strconv.ParseInt(v, 10, 64)
}

// AsIntString returns the parameter as an int64
func (parameterMap *HandlerParameterMap) AsString(id string) (string, error) {
	v, ok := (*parameterMap)[id]

	if !ok {
		return "", &KeyNotFoundError{field: id}
	}

	return v, nil
}
