package service

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/model"
)

// Team is the API representation of a team
type Team struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToModel creates a storage model from the API representation
func (team *Team) ToModel() *model.Team {
	return &model.Team{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
	}
}

// MarshalJSON marshals a JSON string from the API representation
func (team *Team) MarshalJSON() ([]byte, error) {
	return json.Marshal(*team)
}

// NewTeamFromModel creates a HTTP representation of a model team
func NewTeamFromModel(teamModel *model.Team) *Team {
	return &Team{
		ID:          teamModel.ID,
		Name:        teamModel.Name,
		Description: teamModel.Description,
	}
}
