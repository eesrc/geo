package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/sub/manager/event"
	"github.com/eesrc/geo/pkg/sub/manager/topic"
	"github.com/gorilla/mux"
)

func (s *Server) createTeam(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	team, err := validation.ValidateAndGetTeamFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newTeamID, err := validation.CreateTeam(team.ToModel(), s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = validation.SetTeamMember(userProfile.ID, newTeamID, true, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newTeam, err := validation.GetTeam(newTeamID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newTeam.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Team, newTeam.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.CreatedEvent, event.TeamEntity, newTeam.ID),
	)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getTeam(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	team, err := validation.GetTeamFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := team.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateTeam(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	teamID, err := validation.GetTeamID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	teamBody, err := validation.ValidateAndGetTeamFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Ensure ID is not overwritten
	teamBody.ID = teamID

	err = validation.UpdateTeam(teamBody.ToModel(), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updatedTeam, err := validation.GetTeam(teamID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := updatedTeam.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Team, updatedTeam.ID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.UpdatedEvent, event.TeamEntity, updatedTeam.ID),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteTeam(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	teamID, err := validation.GetTeamID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = validation.DeleteTeam(teamID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	s.manager.Publish(
		topic.NewEntityTopic(topic.Team, teamID, topic.LifecycleEvents),
		event.NewLifecycleEvent(event.DeletedEvent, event.TeamEntity, teamID),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listTeams(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	teams, err := validation.ListTeams(userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(teams)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
