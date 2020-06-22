package auth

import "net/http"

// Provider is the provider OAuth interface. Implement this to enable authentication with different providers.
type Provider interface {
	// Enabled tells whether a provider is enabled
	Enabled() bool

	// ServeHTTP adds the different provider endpoints needed for authentication
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	// BasePath is the http base path of the provider
	BasePath() string

	// SetSessions sets the sessions to be used by the provider
	SetSessions(sessions SessionStore)

	// StartSessionChecker will start a go routine to periodically check the sessions for the provider
	// for their validity. Should only be called once.
	StartSessionChecker()

	// SetServerErrorHandleFunc is called with corresponding AuthError when something went wrong
	// during authentication. At this point you must expect the authorization to have
	// failed partially or completely and the user must either try to reauthenticate or the
	// server auth config must be changed
	SetServerErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err AuthError))
}
