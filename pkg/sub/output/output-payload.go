package output

import "github.com/eesrc/geo/pkg/model"

// outputPayload is a simple internal struct for handling tracker movements for a position
type outputPayload struct {
	position  model.Position
	movements []*TrackerMovement
}
