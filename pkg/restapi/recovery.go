package restapi

import (
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	log "github.com/sirupsen/logrus"
)

type recoveryHandler struct {
	internalHandler http.Handler
}

func RecoveryHandler() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		r := &recoveryHandler{internalHandler: handler}
		return r
	}
}

func (handler recoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Unexpected panic occured: %v", err)
			errorResponse := validation.NewErrorResponse(http.StatusServiceUnavailable)
			errorResponse.WriteHTTPError(w)
		}
	}()

	handler.internalHandler.ServeHTTP(w, req)
}
