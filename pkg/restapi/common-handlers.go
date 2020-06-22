package restapi

import (
	"net/http"

	"github.com/eesrc/geo/pkg/restapi/validation"
	log "github.com/sirupsen/logrus"
)

// handleError is a generic handler for handling both validation and store errors
func handleError(err error, w http.ResponseWriter, logger *log.Entry) {
	if err, ok := err.(*validation.Error); ok {
		err.ErrorResponse.WriteHTTPError(w)
		logger.WithError(err).Info("Validation error")
		return
	}

	logger.WithError(err).Error("Store transaction error")
	validation.NewErrorResponse(http.StatusServiceUnavailable).WriteHTTPError(w)
}
