package output

import (
	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/sub"
)

// TrackerMovement represents the last movements relative to a shape
type TrackerMovement struct {
	lastMovements  sub.MovementList
	trackerID      int64
	shapeID        int64
	lastPositionID int64
}

// Update updates the shapeMovement based on inside param and position
func (trackerMovement *TrackerMovement) Update(inside bool, position model.Position) {
	// Get diff
	newMovements := trackerMovement.DiffMovement(inside)

	// Update model
	trackerMovement.lastMovements = newMovements
	trackerMovement.lastPositionID = position.ID
}

// DiffMovement returns a MovementTypes based on param which represents that
// an entity is now inside shape or not
func (trackerMovement *TrackerMovement) DiffMovement(inside bool) sub.MovementList {
	if !trackerMovement.lastMovements.Contains(sub.Inside) && inside {
		return sub.MovementList{sub.Entered, sub.Inside}
	}

	if trackerMovement.lastMovements.Contains(sub.Inside) && !inside {
		return sub.MovementList{sub.Exited, sub.Outside}
	}

	if inside {
		return sub.MovementList{sub.Inside}
	}

	return sub.MovementList{sub.Outside}
}

// NewTrackerMovementFromModel create a TrackerMovement struct based on model input
func NewTrackerMovementFromModel(movement *model.TrackerMovement) *TrackerMovement {
	var trackerMovement = TrackerMovement{
		lastPositionID: movement.PositionID,
		lastMovements:  sub.NewMovementTypeFromModel(movement.Movements),
		shapeID:        movement.ShapeID,
		trackerID:      movement.TrackerID,
	}

	return &trackerMovement
}

// NewTrackerMovementListFromMdel creates a TrackerMovement list from a model TrackerMovement list
func NewTrackerMovementListFromModel(trackerMovementList []model.TrackerMovement) []*TrackerMovement {
	trackerMovements := make([]*TrackerMovement, 0)
	for _, trackerMovement := range trackerMovementList {
		trackerMovements = append(trackerMovements, NewTrackerMovementFromModel(&trackerMovement))
	}

	return trackerMovements
}
