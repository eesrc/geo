package service

import (
	"encoding/json"
	"time"

	"github.com/eesrc/geo/pkg/model"
)

// Position is the API representation of a position
type Position struct {
	ID        int64    `json:"id"`
	TrackerID int64    `json:"trackerId"`
	Timestamp *int64   `json:"timestamp"`
	Lat       *float64 `json:"lat"`
	Long      *float64 `json:"lng"`
	Alt       *float64 `json:"alt"`
	Heading   *float64 `json:"heading"`
	Speed     *float64 `json:"speed"`
	Payload   []byte   `json:"payload"`
	Precision *float64 `json:"precision"`
}

// ToModel creates a storage model from the API representation
func (position *Position) ToModel() *model.Position {
	var timestamp int64

	// Check if timestamp is set. If yes, we assume it's provided in ms. Otherwise set with ns precision.
	if position.Timestamp != nil {
		timestamp = milliToNanoSeconds(*position.Timestamp)
	} else {
		timestamp = time.Now().UnixNano()
	}

	return &model.Position{
		ID:        position.ID,
		TrackerID: position.TrackerID,
		Timestamp: timestamp,
		Lat:       *position.Lat,
		Lon:       *position.Long,
		Alt:       *position.Alt,
		Heading:   *position.Heading,
		Speed:     *position.Speed,
		Payload:   position.Payload,
		Precision: *position.Precision,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (position *Position) MarshalJSON() ([]byte, error) {
	return json.Marshal(*position)
}

// NewPositionFromModel creates a HTTP representation of a model collection
func NewPositionFromModel(positionModel *model.Position) *Position {
	lat := positionModel.Lat
	long := positionModel.Lon
	alt := positionModel.Alt
	precision := positionModel.Precision
	heading := positionModel.Heading
	speed := positionModel.Speed

	timestamp := nanoToMilliSeconds(positionModel.Timestamp)

	return &Position{
		ID:        positionModel.ID,
		TrackerID: positionModel.TrackerID,
		Timestamp: &timestamp,
		Lat:       &lat,
		Long:      &long,
		Alt:       &alt,
		Heading:   &heading,
		Speed:     &speed,
		Payload:   positionModel.Payload,
		Precision: &precision,
	}
}

func nanoToMilliSeconds(nano int64) int64 {
	return nano / int64(time.Millisecond)
}

func milliToNanoSeconds(milli int64) int64 {
	return milli * int64(time.Millisecond)
}

// NewPosition returns a Position with default params
func NewPosition() Position {
	// Optional parameters
	alt := 0.0
	heading := 0.0
	speed := 0.0

	return Position{
		Alt:     &alt,
		Heading: &heading,
		Speed:   &speed,
	}
}
