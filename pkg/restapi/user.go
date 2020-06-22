package restapi

import (
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/restapi/validation"
)

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	userProfile := s.UserFromRequest(r)

	jsonBytes, err := service.NewUserFromModel(userProfile).MarshalJSON()
	if err != nil {
		validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}
