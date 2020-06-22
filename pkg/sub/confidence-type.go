package sub

import (
	"errors"

	"github.com/eesrc/geo/pkg/model"
)

// ConfidenceType is the type of movement subscription
type ConfidenceType string

// ConfidenceList is a list of movement types
type ConfidenceList []ConfidenceType

const (
	// ConfidenceLow is the low confidence of precision [0,.5]
	ConfidenceLow ConfidenceType = "low"
	// ConfidenceMedium is the medium confidence of precision (.5,.75)
	ConfidenceMedium ConfidenceType = "medium"
	// ConfidenceHigh is the high confidence of precision [.75, 1]
	ConfidenceHigh ConfidenceType = "high"
)

// ValidConfidenceTypes is a list of valid movement types
var ValidConfidenceTypes = []ConfidenceType{ConfidenceLow, ConfidenceMedium, ConfidenceHigh}

func (confidenceList *ConfidenceList) ToModel() model.ConfidenceList {
	var modelConfidenceList = model.ConfidenceList{}
	for _, confidence := range *confidenceList {
		modelConfidenceList = append(modelConfidenceList, string(confidence))
	}

	return modelConfidenceList
}

func (confidenceList *ConfidenceList) Contains(confidenceCandidate ConfidenceType) bool {
	for _, confidence := range *confidenceList {
		if confidence == confidenceCandidate {
			return true
		}
	}

	return false
}

// NewConfidenceListFromModel
func NewConfidenceListFromModel(confidenceList model.ConfidenceList) ConfidenceList {
	var newConfidenceList = ConfidenceList{}

	for _, movement := range confidenceList {
		newConfidenceList = append(newConfidenceList, ConfidenceType(movement))
	}

	return newConfidenceList
}

func NewConfidenceFromFloat(precision float64) (ConfidenceType, error) {
	if precision >= 0 && precision < 0.5 {
		return ConfidenceLow, nil
	}
	if precision > 0.5 && precision < 0.75 {
		return ConfidenceMedium, nil
	}
	if precision > 0.75 {
		return ConfidenceHigh, nil
	}

	return "", errors.New("The confidence needs to be be [0,1]")
}
