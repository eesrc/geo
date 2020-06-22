package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/store/errors"
)

// GetTeamID returns a teamID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetTeamID(handlerParams HandlerParameterMap) (int64, error) {
	teamID, err := handlerParams.TeamID()
	if err != nil {
		return -1, newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("teamId", fmt.Sprintf("The team id '%s' is malformed", handlerParams["teamID"])),
		))
	}

	return teamID, nil
}

// GetTeamFromHandlerParams validates team ID (if) found in params and returns a team from store.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetTeamFromHandlerParams(handlerParams HandlerParameterMap, userId int64, store store.Store) (*service.Team, error) {
	teamID, err := GetTeamID(handlerParams)
	if err != nil {
		return &service.Team{}, err
	}

	team, err := GetTeam(teamID, userId, store)
	if err != nil {
		return &service.Team{}, err
	}

	return team, nil
}

// GetTeam validates team (if) found in params and returns a team from store.
// Returns a validation error containing an ErrorResponse based on what went wrong
func GetTeam(teamID int64, userID int64, store store.Store) (*service.Team, error) {
	team, err := store.GetTeamByUserID(teamID, userID)
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return &service.Team{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("teamId", fmt.Sprintf("The team with id '%d' might not exist", teamID)),
				))
			}
		}

		return &service.Team{}, err
	}

	return service.NewTeamFromModel(team), nil
}

// CreateTeam tries to create a team and returns a regular error
// if the creation fails
func CreateTeam(team *model.Team, store store.Store) (int64, error) {
	return store.CreateTeam(team)
}

// SetTeamMember sets a user - team relation and returns a regular error if the connection fails
func SetTeamMember(userID int64, teamID int64, admin bool, store store.Store) error {
	return store.SetTeamMember(userID, teamID, admin)
}

// UpdateTeam tries to update a team and returns a validation error or regular error
// if the update fails
func UpdateTeam(team *model.Team, userID int64, store store.Store) error {
	err := store.UpdateTeam(team, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("team.id", fmt.Sprintf("The collection id '%d' might not exist", team.ID)),
				))
			}
		}
	}

	return err
}

// DeleteTeam tries to delete a team and returns a validation error or regular error
// if the delete fails
func DeleteTeam(teamID int64, userID int64, store store.Store) error {
	err := store.DeleteTeam(teamID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("teamId", fmt.Sprintf("The team with id '%d' might not exist", teamID)),
				))
			}
		}
	}

	return err
}

// ListTeams lists teams based on userID and returns a store error if the list fails
func ListTeams(userID int64, filterParams FilterParams, store store.Store) ([]*service.Team, error) {
	collections, err := store.ListTeamsByUserID(userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.Team{}, err
	}

	var collectionList []*service.Team = make([]*service.Team, len(collections))

	for i, collection := range collections {
		collectionList[i] = service.NewTeamFromModel(&collection)
	}

	return collectionList, nil
}

// ValidateAndGetTeamFromBody retrieves a team from given body and decodes it.
// Returns a validation error containing an ErrorResponse if something went wrong
func ValidateAndGetTeamFromBody(body io.ReadCloser) (*service.Team, error) {
	jsonDecoder := json.NewDecoder(body)
	var team service.Team
	err := jsonDecoder.Decode(&team)
	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &team, getUnmarshalError(err)
		}

		return &team, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("team", "You need to provide a valid team object"),
			),
		)
	}

	return &team, nil
}
