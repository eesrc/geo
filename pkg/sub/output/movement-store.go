package output

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store"
)

// movementStore is a simple aggregator of movements so writing the movements to the DB doesn't
// lag and can be inserted through a transaction
type movementStore struct {
	channel chan model.TrackerMovement
	store   store.Store
}

const maxMovements = 1500
const debounceTimeMS = 500 * time.Millisecond

func (buffer *movementStore) storeMovement(newMovement *model.TrackerMovement) {
	buffer.channel <- *newMovement
}

func newMovementStore(store store.Store) movementStore {
	movementChannel := make(chan model.TrackerMovement, maxMovements)
	newBuffer := movementStore{
		channel: movementChannel,
		store:   store,
	}

	go func() {
		// Create timer and stop it immedietly as we don't want to run until we get a movement
		timer := time.NewTimer(debounceTimeMS)
		timer.Stop()

		movementSync := &sync.Mutex{}
		movementsToBeStored := make([]model.TrackerMovement, 0)

		for {
			select {
			case movement := <-movementChannel:
				// New movement input, add to array to be stored and reset timer
				movementSync.Lock()
				movementsToBeStored = append(movementsToBeStored, movement)
				timer.Reset(debounceTimeMS)
				movementSync.Unlock()
			case <-timer.C:
				movementSync.Lock()
				// The timer has timed out, triggering a storage of available movements
				if len(movementsToBeStored) > 0 {
					err := store.InsertMovements(movementsToBeStored)
					if err != nil {
						log.Error("Failed to update movements")
					}
					movementsToBeStored = nil
				}
				movementSync.Unlock()
			}
		}
	}()
	return newBuffer
}
