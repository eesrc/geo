package sub

import (
	"github.com/eesrc/geo/pkg/model"
)

// MovementType is the type of movement subscription
type MovementType string

// MovementList is a list of movement types
type MovementList []MovementType

// Contains checks wheter a movementTypeCandidate is contained within MovementTypes
func (movementTypes *MovementList) Contains(movementCandidate MovementType) bool {
	for _, movementType := range *movementTypes {
		if movementType == movementCandidate {
			return true
		}
	}
	return false
}

// ToStringSlice returns a list of strings based on the movementTypes
func (movementTypes *MovementList) ToStringSlice() []string {
	stringSlice := make([]string, 0)

	for _, movement := range *movementTypes {
		stringSlice = append(stringSlice, string(movement))
	}

	return stringSlice
}

// ContainsAny checks wheter any of given movementTypeCandidates is contained within MovementTypes
func (movementTypes *MovementList) ContainsAny(movementCandidates MovementList) bool {
	for _, movementType := range *movementTypes {
		for _, movementCandidate := range movementCandidates {
			if movementType == movementCandidate {
				return true
			}
		}
	}
	return false
}

// ToModel returns a model representation of the movement list
func (movementTypes *MovementList) ToModel() model.MovementList {
	var modelMovements = model.MovementList{}
	for _, triggerMovement := range *movementTypes {
		modelMovements = append(modelMovements, string(triggerMovement))
	}

	return modelMovements
}

// NewMovementList returns a list of movements based on an inside bool param
func NewMovementList(inside bool) MovementList {
	if inside {
		return MovementList{Entered, Inside}
	}

	return MovementList{}
}

// NewMovementTypeFromModel returns a MovementList from model
func NewMovementTypeFromModel(movements model.MovementList) MovementList {
	var triggerMovements = []MovementType{}
	for _, movement := range movements {
		triggerMovements = append(triggerMovements, MovementType(movement))
	}

	return triggerMovements
}

const (
	// Entered is the state when a tracker has entered a subscribed shape
	Entered MovementType = "entered"
	// Inside is the state when a tracker is inside a subscribed shape
	Inside MovementType = "inside"
	// Exited is the state when a tracker has exited a subscribed shape
	Exited MovementType = "exited"
	// Outside is the state when a tracker has exited a subscribed shape
	Outside MovementType = "outside"
)

// ValidMovementTypes is a list of valid movement types
var ValidMovementTypes = []MovementType{Entered, Inside, Exited, Outside}
