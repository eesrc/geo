package output

import (
	"sync"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/sub"
	"github.com/eesrc/geo/pkg/tria/geometry"
)

type movementIndex struct {
	mutex    *sync.Mutex
	trackers map[int64][]*TrackerMovement
}

func newMovementIndex() movementIndex {
	return movementIndex{
		mutex:    &sync.Mutex{},
		trackers: make(map[int64][]*TrackerMovement),
	}
}

func (movementIndex *movementIndex) addMovements(movements []*TrackerMovement) {
	movementIndex.mutex.Lock()
	defer movementIndex.mutex.Unlock()

	for _, newTrackerMovement := range movements {
		var trackerMovements = movementIndex.trackers[newTrackerMovement.trackerID]

		// First movement registered
		if trackerMovements == nil {
			movementIndex.trackers[newTrackerMovement.trackerID] = []*TrackerMovement{newTrackerMovement}
			continue
		}

		// Check if the movement is replacing any other movement
		for idx, trackerMovement := range trackerMovements {
			if trackerMovement.shapeID == newTrackerMovement.shapeID {
				trackerMovements[idx] = newTrackerMovement
				continue
			}
		}

		// Not found and not first, add to list
		movementIndex.trackers[newTrackerMovement.trackerID] = append(trackerMovements, newTrackerMovement)
	}
}

func (movementIndex *movementIndex) setAndDiffMovement(position model.Position, shapes []geometry.Shape) []*TrackerMovement {
	movementIndex.mutex.Lock()
	defer movementIndex.mutex.Unlock()

	trackerMovementList, ok := movementIndex.trackers[position.TrackerID]

	// No tracker movements registered before. Add new movement list for tracker
	if !ok || len(trackerMovementList) == 0 {
		newTrackerMovementList := make([]*TrackerMovement, len(shapes))

		for i, shape := range shapes {
			newTrackerMovementList[i] = &TrackerMovement{
				shapeID:        shape.GetID(),
				trackerID:      position.TrackerID,
				lastPositionID: position.ID,
				lastMovements:  sub.NewMovementList(true),
			}
		}

		movementIndex.trackers[position.TrackerID] = newTrackerMovementList

		return newTrackerMovementList
	}

	newTrackerMovementList := make([]*TrackerMovement, 0)

	// Tracker is outside of all shapes, iterate and update all movement as outside.
	if len(shapes) == 0 {
		for _, trackerMovement := range trackerMovementList {
			trackerMovement.Update(false, position)

			if len(trackerMovement.lastMovements) > 0 {
				newTrackerMovementList = append(newTrackerMovementList, trackerMovement)
			}
		}

		movementIndex.trackers[position.TrackerID] = newTrackerMovementList
		return movementIndex.trackers[position.TrackerID]
	}

	// Do a full diff of available tracker movements towards given shapes
	for _, trackerMovement := range trackerMovementList {
		found := false

		for _, shape := range shapes {
			if trackerMovement.shapeID == shape.GetID() {
				found = true

				trackerMovement.Update(true, position)

				if len(trackerMovement.lastMovements) > 0 {
					newTrackerMovementList = append(newTrackerMovementList, trackerMovement)
				}
			}
		}

		if !found {
			trackerMovement.Update(false, position)

			if len(trackerMovement.lastMovements) > 0 {
				newTrackerMovementList = append(newTrackerMovementList, trackerMovement)
			}
		}
	}

	// Add new movements for non-existing movements based on given shapes
	for _, shape := range shapes {
		found := false

		for _, trackerMovement := range trackerMovementList {
			if trackerMovement.shapeID == shape.GetID() {
				found = true
			}
		}

		if !found {
			newTrackerMovementList = append(newTrackerMovementList, &TrackerMovement{
				shapeID:        shape.GetID(),
				trackerID:      position.TrackerID,
				lastMovements:  sub.NewMovementList(true),
				lastPositionID: position.ID,
			})
		}
	}

	movementIndex.trackers[position.TrackerID] = newTrackerMovementList

	return newTrackerMovementList
}
