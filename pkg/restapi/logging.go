package restapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const uuidContextKey = "uuid"

// RequestLogger returns a populated logger with an unique UUID for the request
// along with user info if present
func (s *Server) RequestLogger(r *http.Request) *log.Entry {
	requestLogger := getLoggingInstance(r)

	userProfile := s.UserFromRequest(r)
	if userProfile != nil {
		requestLogger = requestLogger.WithField("userId", userProfile.ID)
	}

	authType := s.AuthTypeKeyFromRequest(r)
	if authType != "" {
		requestLogger = requestLogger.WithField("authType", authType)
	}

	return requestLogger
}

func getLoggingInstance(r *http.Request) *log.Entry {
	return log.WithField("uuid", getUUIDFromContext(r.Context()))
}

func getUUIDFromContext(ctx context.Context) uuid.UUID {
	contextUUID := ctx.Value(uuidContextKey)

	if contextUUID == nil {
		log.Warn("Could not find a UUID on context, generating a new one but not persisting.")
		return uuid.New()
	}

	return contextUUID.(uuid.UUID)
}

func UUIDWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if uuid := ctx.Value(uuidContextKey); uuid != nil {
			log.Info("Context already has an UUID, returning clean serve")
			next.ServeHTTP(w, r)
			return
		}

		log.Debug("Adding new UUID to request")
		next.ServeHTTP(
			w,
			r.WithContext(
				context.WithValue(
					ctx,
					uuidContextKey,
					uuid.New(),
				),
			),
		)
	})
}
