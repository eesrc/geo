package providers

import (
	"strings"

	"golang.org/x/oauth2/github"
)

const (
	// githubUserURL is the url to fetch the github user profile from
	githubUserProfileURL = "https://api.github.com/user"
	// githubTokenCheckURL is the URL used to check the validity of the github token
	githubTokenCheckURL = "https://api.github.com/applications/%s/token"
)

// A simplified Connect specific configuration
type GithubConfig struct {
	// Enabled determines whether the auth is enabled
	Enabled bool `param:"desc=Auth type enabled;default=false"`

	// ClientID is the Github client ID
	ClientID string `param:"desc=Client ID"`
	// ClientSecret is the Github secret
	ClientSecret string `param:"desc=Client secret"`

	// Scopes is a comma separated list of scopes to request
	Scopes string `param:"desc=Scopes to request towards Connect;default=user:read,user:email"`

	// AuthBasePath is the base path for the server github endpoint. It will be used when adding
	// paths to the router.
	AuthBasePath string `param:"desc=The base path for the provider;default=/github"`

	// CallbackURL is the callback URL to be used during the OAuth2 flow
	CallbackURL string `param:"desc=Callback URL;default=http://localhost:8080/github/oauth2callback"`
	// LoginSuccessURL is the URL to redirect to when a successful login has occured
	LoginSuccessURL string `param:"desc=Login success redirect URL;default=/"`
	// LogoutSuccessURL is the URL to redirect to when a successful logout has occured
	LogoutSuccessURL string `param:"desc=Logout success redirect URL;default=/"`

	// SecureCookie determines whether the session cookie is set to be secure (ie, https only)
	SecureCookie bool `param:"desc=Determines the session cookie is secure (HTTPS) only;default=false"`
}

func NewConfigFromGithub(githubConfig GithubConfig) Config {
	config := NewFromConfig(Config{
		Enabled:          githubConfig.Enabled,
		ClientID:         githubConfig.ClientID,
		ClientSecret:     githubConfig.ClientSecret,
		Scopes:           strings.Split(githubConfig.Scopes, ","),
		AuthBasePath:     githubConfig.AuthBasePath,
		CallbackURL:      githubConfig.CallbackURL,
		LoginSuccessURL:  githubConfig.LoginSuccessURL,
		LogoutSuccessURL: githubConfig.LogoutSuccessURL,
		SecureCookie:     githubConfig.SecureCookie,

		AuthEndpointURL:  github.Endpoint.AuthURL,
		TokenEndpointURL: github.Endpoint.TokenURL,
		UserEndpointURL:  githubUserProfileURL,
	})
	return config
}
