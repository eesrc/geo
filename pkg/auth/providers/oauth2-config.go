package providers

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/auth"
)

var (
	defaultAuthBasePath     = "/auth"
	defaultLoginSuccessURL  = "/"
	defaultLogoutSuccessURL = "/"
	defaultCallbackURL      = defaultAuthBasePath + "/oauth2callback"

	defaultScopes = []string{"openid", "profile", "email"}

	defaultAuthEndpointURL  = "/"
	defaultTokenEndpointURL = "/"

	defaultErrorHandler = func(w http.ResponseWriter, r *http.Request, err auth.AuthError) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(err.Error()))
	}
	defaultProfileHandler = func(bytes []byte) auth.Profile {
		profile := auth.Profile{}

		_ = profile.Scan(bytes)

		return profile
	}
	defaultSessionChecker = func(auth.SessionStore, Config) {
		log.Warn("No configured sessionchecker. Authorization may be updated.")
	}
	defaultSecureCookie = true
)

// Config is the generic OAuth2 config for an OAuth2 provider
type Config struct {
	// Enabled determines whether the auth is enabled
	Enabled bool `param:"desc=Auth type enabled;default=false"`

	// ClientID is the OAuth2 provider client ID
	ClientID string `param:"desc=Client ID"`
	// ClientSecret is the OAuth2 provider secret
	ClientSecret string `param:"desc=Client secret"`

	// Scopes are a list of scopes to be used during the auth process
	Scopes []string

	// AuthBasePath is the base path for the provider. It will be used when adding
	// paths to the router.
	AuthBasePath string `param:"desc=The base path for the provider"`

	// CallbackURL is the callback URL to be used during the OAuth2 flow
	CallbackURL string `param:"desc=Callback URL;default=/"`
	// LoginSuccessURL is the URL to redirect to when a successful login has occured
	LoginSuccessURL string `param:"desc=Login success redirect URL;default=/"`
	// LogoutSuccessURL is the URL to redirect to when a successful logout has occured
	LogoutSuccessURL string `param:"desc=Logout success redirect URL;default=/"`

	// AuthEndpointURL is the auth endpoint for the OAuth2 provider
	AuthEndpointURL string `param:"desc=The OAuth2 auth URL endpoint;default=/"`
	// TokenEndpointURL is the token endpoint for the OAuth2 provider
	TokenEndpointURL string `param:"desc=The Oauth2 token URL endpoint;default=/"`
	// UserEndpointURL is the user endpoint where the provider will get information about the user
	UserEndpointURL string `param:"desc=The URL to get user information from using the authenticated token;default=/"`

	// SecureCookie determines whether the session cookie is set to be secure (ie, https only)
	SecureCookie bool `param:"desc=Determines the session cookie is secure (HTTPS) only"`
}

// NewFromConfig returns a populated Config object with defaults
func NewFromConfig(overrideConfig Config) Config {
	config := Config{
		Enabled:          false,
		AuthBasePath:     defaultAuthBasePath,
		LoginSuccessURL:  defaultLoginSuccessURL,
		LogoutSuccessURL: defaultLogoutSuccessURL,
		Scopes:           defaultScopes,
		AuthEndpointURL:  defaultAuthEndpointURL,
		TokenEndpointURL: defaultTokenEndpointURL,
		CallbackURL:      defaultCallbackURL,
		SecureCookie:     defaultSecureCookie,
	}

	if overrideConfig.Enabled != config.Enabled {
		config.Enabled = overrideConfig.Enabled
	}

	if overrideConfig.AuthBasePath != "" {
		config.AuthBasePath = overrideConfig.AuthBasePath
	}

	if overrideConfig.ClientID != "" {
		config.ClientID = overrideConfig.ClientID
	}

	if overrideConfig.ClientSecret != "" {
		config.ClientSecret = overrideConfig.ClientSecret
	}

	if overrideConfig.Scopes != nil {
		config.Scopes = overrideConfig.Scopes
	}

	if overrideConfig.LoginSuccessURL != "" {
		config.LoginSuccessURL = overrideConfig.LoginSuccessURL
	}

	if overrideConfig.LogoutSuccessURL != "" {
		config.LogoutSuccessURL = overrideConfig.LogoutSuccessURL
	}

	if overrideConfig.AuthEndpointURL != "" {
		config.AuthEndpointURL = overrideConfig.AuthEndpointURL
	}

	if overrideConfig.TokenEndpointURL != "" {
		config.TokenEndpointURL = overrideConfig.TokenEndpointURL
	}

	if overrideConfig.UserEndpointURL != "" {
		config.UserEndpointURL = overrideConfig.UserEndpointURL
	}

	if overrideConfig.CallbackURL != "" {
		config.CallbackURL = overrideConfig.CallbackURL
	}

	if overrideConfig.SecureCookie != config.SecureCookie {
		config.SecureCookie = overrideConfig.SecureCookie
	}

	return config
}
