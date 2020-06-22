package providers

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/auth"
	"golang.org/x/oauth2"
)

type OAuth2Provider struct {
	Config     Config // Config is the Provider configuration
	sessions   auth.SessionStore
	authConfig oauth2.Config

	userToProfileFunc     func([]byte) auth.Profile
	serverErrorHandleFunc func(w http.ResponseWriter, r *http.Request, err auth.AuthError)
	sessionChecker        func(auth.SessionStore, Config)
}

// Handler returns a handler for the (local) GitHub resource
func (p *OAuth2Provider) Handler() http.Handler {
	return p
}

const (
	loginPath    = "/login"
	callbackPath = "/oauth2callback"
	logoutPath   = "/logout"
)

// ServeHTTP is a http.HandleFunc to be attached to your route of choice. It will add a
// login path, a callback path and a logout path. If it receives an unknown request, it
// will use the ServerErrorHandleFunc provided in the Authenticator configuration.
func (p *OAuth2Provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, loginPath) {
		// Generic handling of OAuth2 login flow
		handleOAuth2Login(w, r, p)
		return
	}

	if strings.HasSuffix(r.URL.Path, callbackPath) {
		// Generic handling of OAuth2 code callback
		handleOAuth2Callback(w, r, p)
		return
	}
	if strings.HasSuffix(r.URL.Path, logoutPath) {
		// Generic handling of OAuth2 logout flow
		handleOauth2Logout(w, r, p)
		return
	}
	log.Warnf("Don't know how to handle path '%s' in OAuth2-provider", r.URL.Path)
	p.serverErrorHandleFunc(w, r, auth.UnknownMethodError)

}

func (p *OAuth2Provider) Enabled() bool {
	return p.Config.Enabled
}

func (p *OAuth2Provider) SetSessions(sessions auth.SessionStore) {
	p.sessions = sessions
}

func (p *OAuth2Provider) SetServerErrorHandlerFunc(serverErrorFunc func(w http.ResponseWriter, r *http.Request, err auth.AuthError)) {
	p.serverErrorHandleFunc = serverErrorFunc
}

// StartSessionChecker launches a profile checker goroutine
func (p *OAuth2Provider) StartSessionChecker() {
	go func() {
		p.sessionChecker(p.sessions, p.Config)
	}()
}

// BasePath returns the OAuth2 base path to be used when adding provider routes
func (p *OAuth2Provider) BasePath() string {
	return p.Config.AuthBasePath
}
