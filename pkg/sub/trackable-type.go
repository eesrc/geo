package sub

import (
	"database/sql/driver"

	"github.com/eesrc/geo/pkg/serializing"
)

// TrackableType is a trackable type
type TrackableType string

// Value implements SQL value driver
func (trackableType TrackableType) Value() (driver.Value, error) {
	return serializing.ValueJSON(trackableType)
}

// Scan implements SQL scan driver
func (trackableType *TrackableType) Scan(src interface{}) error {
	return serializing.ScanJSON(trackableType, src)
}

const (
	// Tracker represent a tracker as a trackable type
	Tracker TrackableType = "tracker"
	// Collection represents a collection as a trackable type
	Collection TrackableType = "collection"
)

// ValidTrackableTypes is a list of valid trackable types
var ValidTrackableTypes = []TrackableType{Tracker, Collection}
