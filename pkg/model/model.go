package model

import (
	"database/sql/driver"
	"time"

	"github.com/eesrc/geo/pkg/serializing"
	"github.com/eesrc/geo/pkg/tria/geometry"
)

// User represents a user
type User struct {
	ID            int64
	Name          string
	Email         string
	EmailVerified bool
	Phone         string
	PhoneVerified bool
	Deleted       bool
	Admin         bool
	Created       time.Time
	GithubID      string
	ConnectID     string
}

// Token represents API tokens
type Token struct {
	Token     string
	Resource  string
	UserID    int64
	PermWrite bool
	Created   time.Time
}

// Team represents a team
type Team struct {
	ID          int64
	Name        string
	Description string
}

// Collection is a collection of trackers
type Collection struct {
	ID          int64
	TeamID      int64
	Name        string
	Description string
}

// Tracker represents a device that is being tracked
type Tracker struct {
	ID           int64
	CollectionID int64
	Name         string
	Description  string
}

// ShapeCollection represents a collection of spatial shapes as a polygon or circle
type ShapeCollection struct {
	ID          int64
	TeamID      int64
	Name        string
	Description string
}

// Shape represents a shape, part of a ShapeCollection
type Shape struct {
	ID                int64
	ShapeCollectionID int64
	Name              string
	Properties        geometry.ShapeProperties
	Shape             geometry.Shape
}

// ShapeFeature represents a shape, part of a ShapeCollection, containing both raw and triangulized data
type ShapeFeature struct {
	ID                int64
	ShapeCollectionID int64
	Name              string
	Properties        geometry.ShapeProperties
	Shape             geometry.Shape
}

// Position represents a datapoint from a tracker which may optionally
// carry a payload.
type Position struct {
	ID        int64
	TrackerID int64
	Timestamp int64
	Lat       float64
	Lon       float64
	Alt       float64
	Heading   float64
	Speed     float64
	Payload   []byte
	Precision float64
}

// TrackerMovement contains information where the position was last
type TrackerMovement struct {
	TrackerID      int64
	SubscriptionID int64
	ShapeID        int64
	PositionID     int64
	Movements      MovementList
}

// Subscription represents a subscription between a trackable entity and an output
type Subscription struct {
	ID                int64
	TeamID            int64
	Name              string
	Description       string
	Active            bool
	Output            string
	OutputConfig      OutputConfig
	Types             MovementList
	Confidences       ConfidenceList
	ShapeCollectionID int64
	// TrackableType is which type of trackable we're talking about, typically a tracker
	// or a collection of trackers
	TrackableType string
	// TrackableID is the ID of trackable within its domain, either tracker og collection
	TrackableID int64
}

// GeoSubscription represents an aggregated struct containing both the subscription details,
// shape collection, subscription shapes and last tracker movements, all of which is needed to
// activate a full subscription between a Subscription and a trackable.
type GeoSubscription struct {
	Subscription     Subscription
	ShapeCollection  ShapeCollection
	TrackerMovements []TrackerMovement
	Shapes           []Shape
}

// OutputConfig is a generic mapping of config for an output
type OutputConfig map[string]interface{}

// Value implements SQL value driver
func (config OutputConfig) Value() (driver.Value, error) {
	return serializing.ValueJSON(config)
}

// Scan implements SQL scan driver
func (config *OutputConfig) Scan(src interface{}) error {
	return serializing.ScanJSON(config, src)
}

// OutputConfig is a generic mapping of config for an output
type MovementList []string

// Value implements SQL value driver
func (list MovementList) Value() (driver.Value, error) {
	return serializing.ValueJSON(list)
}

// Scan implements SQL scan driver
func (list *MovementList) Scan(src interface{}) error {
	return serializing.ScanJSON(list, src)
}

// OutputConfig is a generic mapping of config for an output
type ConfidenceList []string

// Value implements SQL value driver
func (list ConfidenceList) Value() (driver.Value, error) {
	return serializing.ValueJSON(list)
}

// Scan implements SQL scan driver
func (list *ConfidenceList) Scan(src interface{}) error {
	return serializing.ScanJSON(list, src)
}
