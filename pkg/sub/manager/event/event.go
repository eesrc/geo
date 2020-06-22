package event

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

// EventType defines which type of event the event consist of
type EventType string

const (
	// LifeCycle represents the lifecycle event type
	LifeCycle EventType = "lifecycle"
	// Position represents the position event type
	Position EventType = "position"
	// Subscription represents the subscription event type
	Subscription EventType = "subscription"
)

// DecodeEvent tries to decode a received message to either a PositionEvent, LifeycleEvent or SubscriptionEvent
func DecodeEvent(data interface{}) (interface{}, error) {
	var event map[string]interface{} = make(map[string]interface{})
	bytes, ok := data.([]byte)

	if !ok {
		return event, fmt.Errorf("Provided type is not a bytestream")
	}

	err := json.Unmarshal(bytes, &event)
	if err != nil {
		return event, err
	}

	if event["type"] == nil || reflect.TypeOf(event["type"]).Name() != "string" {
		return event, fmt.Errorf("The given data does not contain any type information")
	}

	// Go switch on available types and remarshall to fit values
	switch EventType(event["type"].(string)) {
	case LifeCycle:
		var lifecycleEvent LifeCycleEvent
		config := mapstructure.DecoderConfig{
			DecodeHook: mapperHook,
			Result:     &lifecycleEvent,
		}
		decoder, _ := mapstructure.NewDecoder(&config)

		err = decoder.Decode(event)
		return lifecycleEvent, err
	case Position:
		var positionEvent PositionEvent
		config := mapstructure.DecoderConfig{
			DecodeHook: mapperHook,
			Result:     &positionEvent,
		}
		decoder, _ := mapstructure.NewDecoder(&config)

		err = decoder.Decode(event)
		return positionEvent, err
	case Subscription:
		var subscriptionEvent SubscriptionEvent
		config := mapstructure.DecoderConfig{
			DecodeHook: mapperHook,
			Result:     &subscriptionEvent,
		}
		decoder, _ := mapstructure.NewDecoder(&config)

		err = decoder.Decode(event)
		return subscriptionEvent, err
	default:
		return event, fmt.Errorf("Couldn't match any type to type '%s'", event["type"])
	}
}

// mapperHook is a convenience function which takes care of haphardously lost types when marshaling and unmarshaling
// between JSON model -> map -> JSON model
func mapperHook(fromType reflect.Type, toType reflect.Type, data interface{}) (interface{}, error) {
	// Map payload to byte array
	if toType == reflect.TypeOf([]byte{}) && fromType == reflect.TypeOf("") {
		return []byte(data.(string)), nil
	}

	// Handle string conversion from RFC3339 to Time
	if toType == reflect.TypeOf(time.Time{}) && fromType == reflect.TypeOf("") {
		return time.Parse(time.RFC3339, data.(string))
	}

	return data, nil
}
