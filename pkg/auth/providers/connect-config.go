package providers

import "strings"

const (
	connectAuthorizePath = "/oauth/authorize"
	connectTokenPath     = "/oauth/token"
	connectUserInfoPath  = "/oauth/userinfo"
)

var (
	// This is the default scopes requested from Connect
	defaultConnectScopes = []string{"openid", "profile", "email", "phone"}
)

// A simplified Connect specific configuration
type ConnectConfig struct {
	// Enabled determines whether the auth is enabled
	Enabled bool `param:"desc=Auth type enabled;default=false"`

	// ClientID is the Connect client ID. Defaults to the open test example from Connect.
	ClientID string `param:"desc=Client ID;default=telenordigital-connectexample-web"`
	// ClientSecret is the Connect secret
	ClientSecret string `param:"desc=Client secret"`

	// Scopes is a comma separated list of scopes to request
	Scopes string `param:"desc=Scopes to request towards Connect;default=openid,profile,email,phone"`

	// AuthBasePath is the base path for the server Connect endpoint. It will be used when adding
	// paths to the router.
	AuthBasePath string `param:"desc=The base path for the provider;default=/connect"`

	// CallbackURL is the callback URL to be used during the OAuth2 flow
	CallbackURL string `param:"desc=Callback URL;default=http://localhost:8080/connect/oauth2callback"`
	// LoginSuccessURL is the URL to redirect to when a successful login has occured
	LoginSuccessURL string `param:"desc=Login success redirect URL;default=/"`
	// LogoutSuccessURL is the URL to redirect to when a successful logout has occured
	LogoutSuccessURL string `param:"desc=Logout success redirect URL;default=/"`

	// Host is the Connect host environment to be used, defaults to https://connect.staging.telenordigital.com
	Host string `param:"desc=The full host for the Connect environment to be used;default=https://connect.staging.telenordigital.com"`

	// SecureCookie determines whether the session cookie is set to be secure (ie, https only)
	SecureCookie bool `param:"desc=Determines the session cookie is secure (HTTPS) only;default=false"`
}

func NewConfigFromConnect(connectConfig ConnectConfig) Config {
	config := NewFromConfig(Config{
		Enabled:          connectConfig.Enabled,
		ClientID:         connectConfig.ClientID,
		ClientSecret:     connectConfig.ClientSecret,
		Scopes:           strings.Split(connectConfig.Scopes, ","),
		AuthBasePath:     connectConfig.AuthBasePath,
		CallbackURL:      connectConfig.CallbackURL,
		LoginSuccessURL:  connectConfig.LoginSuccessURL,
		LogoutSuccessURL: connectConfig.LogoutSuccessURL,
		SecureCookie:     connectConfig.SecureCookie,

		AuthEndpointURL:  connectConfig.Host + connectAuthorizePath,
		TokenEndpointURL: connectConfig.Host + connectTokenPath,
		UserEndpointURL:  connectConfig.Host + connectUserInfoPath,
	})
	return config
}
