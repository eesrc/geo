package auth

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Authenticator is the OAuth interface. Create this to enable authentication
type Authenticator struct {
	sessionStore SessionStore
	config       AuthenticatorConfig
	Providers    []Provider
}

func NewAuthenticator(config AuthenticatorConfig) (*Authenticator, error) {
	sessionStore, err := NewSQLSessionStore(config.DBDriver, config.DBConnectionString)
	if err != nil {
		log.Error("Failed to create SQL Session store")
		return nil, err
	}

	return &Authenticator{
		config:       config,
		sessionStore: sessionStore,
	}, nil
}

// Profile reads the user profile from the http.Request cookie
func (a *Authenticator) Profile(w http.ResponseWriter, r *http.Request) (Profile, error) {
	profile, err := getSessionProfileFromCookie(w, r, a.sessionStore)

	if err != nil {
		return Profile{}, err
	}

	return profile, nil
}

// AddProvider adds a provider and provisions the provider with the
// shared sessions store for the authenticator
func (a *Authenticator) AddProvider(provider Provider) {
	// Don't bother to add provider unless it's enabled
	if provider.Enabled() {
		provider.SetSessions(a.sessionStore)
		a.Providers = append(a.Providers, provider)
	}
}

// ServerErrorHandlerFunc is a handler func for handling server errors which might happen during authentication.
func (a *Authenticator) ServerErrorHandlerFunc(serverErrorFunc func(w http.ResponseWriter, r *http.Request, err AuthError)) {
	for _, provider := range a.Providers {
		provider.SetServerErrorHandlerFunc(serverErrorFunc)
	}
}

const (
	// AuthCookieName is the name of the cookie used to store the session ID
	AuthCookieName = "ee_session"
	// ProfileContextKey is the context key for the active Profile
	ProfileContextKey = "ee_context_profile"
)
