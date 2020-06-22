package restapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/gorilla/mux"
)

func (s *Server) createToken(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	tokenBody, err := validation.GetTokenFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Set token userID
	tokenBody.UserID = userProfile.ID

	// Generate new random token, if set it will be overwritten
	err = tokenBody.GenerateToken()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
	}

	tokenBody.Created = time.Now()

	token, err := validation.CreateToken(tokenBody.ToModel(), s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	newToken, err := validation.GetToken(token, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := newToken.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) updateToken(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	tokenID, err := validation.GetTokenID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	tokenBody, err := validation.GetTokenFromBody(r.Body)
	if err != nil {
		handleError(err, w, log)
		return
	}

	// Set token ID from path and set user ID explicitly
	tokenBody.Token = tokenID
	tokenBody.UserID = userProfile.ID

	err = validation.UpdateToken(tokenBody.ToModel(), s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	updatedToken, err := validation.GetToken(tokenID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := updatedToken.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) getToken(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	token, err := validation.GetTokenFromHandlerParams(mux.Vars(r), userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := token.MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

func (s *Server) deleteToken(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	tokenID, err := validation.GetTokenID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	err = validation.DeleteToken(tokenID, userProfile.ID, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listTokens(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	tokens, err := validation.ListTokens(userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(tokens)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
