package output

import (
	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/sub"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/tria/geometry"
	"github.com/eesrc/geo/pkg/tria/index"
)

// GeoSubscription contains everything needed to handle a subscription in memory
type GeoSubscription struct {
	Subscription  model.Subscription
	Index         index.TriaIndex
	MovementIndex movementIndex
	movementStore movementStore
}

// NewGeoSubscription returns an initialized GeoSubscription with given params
func NewGeoSubscription(subscription model.Subscription, index index.TriaIndex, store store.Store) GeoSubscription {
	return GeoSubscription{
		Subscription:  subscription,
		Index:         index,
		MovementIndex: newMovementIndex(),
		movementStore: newMovementStore(store),
	}
}

// NewGeoSubscriptionWithMovements returns an initialized GeoSubscription with a populated movementIndex based on given movements
func NewGeoSubscriptionWithMovements(subscription model.Subscription, index index.TriaIndex, store store.Store, movements []model.TrackerMovement) GeoSubscription {
	movementIndex := newMovementIndex()
	movementIndex.addMovements(NewTrackerMovementListFromModel(movements))

	return GeoSubscription{
		Subscription:  subscription,
		Index:         index,
		MovementIndex: movementIndex,

		movementStore: newMovementStore(store),
	}
}

// NewGeoSubscriptionFromModel creates a new GeoSubscription from a model. This will create both the movement index and general index for lookup
func NewGeoSubscriptionFromModel(geoSubscription model.GeoSubscription, store store.Store) GeoSubscription {
	index := index.NewRTreeIndexFromModel(geoSubscription.Shapes)

	movementIndex := newMovementIndex()
	movementIndex.addMovements(NewTrackerMovementListFromModel(geoSubscription.TrackerMovements))

	return GeoSubscription{
		Subscription:  geoSubscription.Subscription,
		Index:         index,
		MovementIndex: movementIndex,

		movementStore: newMovementStore(store),
	}
}

// FindShapesWhichContainsPoint checks internal index if it contains a position
func (geoSubscription *GeoSubscription) FindShapesWhichContainsPoint(position model.Position) []geometry.Shape {
	return geoSubscription.Index.FindShapesWhichContainsPoint(geometry.Point{X: position.Lon, Y: position.Lat})
}

// ContainsAnyMovements checks if the subscription contains any of the given movement types
func (geoSubscription *GeoSubscription) ContainsAnyMovements(movementTypes sub.MovementList) bool {
	for _, triggerMovement := range geoSubscription.Subscription.Types {
		for _, movement := range movementTypes {
			if triggerMovement == string(movement) {
				return true
			}
		}
	}
	return false
}

// SetAndDiffMovement diffs and updates the movements with the given shapes and returns a list of movements
// based on given input
func (geoSubscription *GeoSubscription) SetAndDiffMovement(position model.Position, shapes []geometry.Shape) []*TrackerMovement {
	diffedMovements := geoSubscription.MovementIndex.setAndDiffMovement(position, shapes)
	for _, movement := range diffedMovements {
		geoSubscription.movementStore.storeMovement(&model.TrackerMovement{
			SubscriptionID: geoSubscription.Subscription.ID,
			TrackerID:      geoSubscription.Subscription.TrackableID,
			ShapeID:        movement.shapeID,
			PositionID:     position.ID,
			Movements:      movement.lastMovements.ToModel(),
		})
	}
	return diffedMovements
}

func (geoSubscription *GeoSubscription) SubscribedToPrecision(precision float64) bool {
	confidence, err := sub.NewConfidenceFromFloat(precision)
	if err != nil {
		return false
	}

	list := sub.NewConfidenceListFromModel(geoSubscription.Subscription.Confidences)
	return list.Contains(confidence)
}

// GetOutputPayloadFromEvent parses the event message and applies the movements to the subscription. The outputPayload
// includes the movements and position. If the Position does not satisfy the subscription parameters, an empty outputPayload
// is returned
func (geoSubscription *GeoSubscription) GetOutputPayloadFromEvent(message interface{}) (outputPayload, error) {
	decodedEvent, err := event.DecodeEvent(message)
	if err != nil {
		return outputPayload{}, err
	}

	position := decodedEvent.(event.PositionEvent).Data.Position

	// If the precision of the position is not within the subscription parameters we must not
	// propagate movements to subscription to avoid false positive notifications
	if !geoSubscription.SubscribedToPrecision(position.Precision) {
		return outputPayload{}, nil
	}

	// Search for shapes and movements for position
	shapes := geoSubscription.FindShapesWhichContainsPoint(position)
	movements := geoSubscription.SetAndDiffMovement(position, shapes)

	return outputPayload{
		position:  position,
		movements: movements,
	}, nil
}
