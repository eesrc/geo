package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/gorilla/mux"
)

func (s *Server) listCollectionSubscriptions(w http.ResponseWriter, r *http.Request) {
	log := s.RequestLogger(r)
	userProfile := s.UserFromRequest(r)

	collectionID, err := validation.GetCollectionID(mux.Vars(r))
	if err != nil {
		handleError(err, w, log)
		return
	}

	filterParams, err := validation.NewFilterParamsFromQueryParams(r.URL.Query())
	if err != nil {
		handleError(err, w, log)
		return
	}

	subscriptions, err := validation.ListSubscriptionsByCollectionID(collectionID, userProfile.ID, filterParams, s.store)
	if err != nil {
		handleError(err, w, log)
		return
	}

	jsonBytes, err := json.Marshal(subscriptions)
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
