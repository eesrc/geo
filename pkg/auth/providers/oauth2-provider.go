package providers

import (
	"github.com/eesrc/geo/pkg/auth"
	"golang.org/x/oauth2"
)

func NewOAuth2Provider(config Config) auth.Provider {
	return &OAuth2Provider{
		Config: config,
		authConfig: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthEndpointURL,
				TokenURL: config.TokenEndpointURL,
			},
			RedirectURL: config.CallbackURL,
		},

		userToProfileFunc:     defaultProfileHandler,
		sessionChecker:        defaultSessionChecker,
		serverErrorHandleFunc: defaultErrorHandler,
	}
}

// NewGitHubProvider creates a GitHub OAuth2 provider instance
func NewGitHubProvider(githubConfig GithubConfig) auth.Provider {
	config := NewConfigFromGithub(githubConfig)

	return &OAuth2Provider{
		Config: config,
		authConfig: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthEndpointURL,
				TokenURL: config.TokenEndpointURL,
			},
			RedirectURL: config.CallbackURL,
		},

		userToProfileFunc:     profileFromGithubUser,
		sessionChecker:        githubSessionChecker,
		serverErrorHandleFunc: defaultErrorHandler,
	}
}

// NewConnectProvider creates a ConnectID OAuth2 provider instance
func NewConnectProvider(connectConfig ConnectConfig) auth.Provider {
	config := NewConfigFromConnect(connectConfig)

	return &OAuth2Provider{
		Config: config,
		authConfig: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       defaultConnectScopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthEndpointURL,
				TokenURL: config.TokenEndpointURL,
			},
			RedirectURL: config.CallbackURL,
		},

		userToProfileFunc:     profileFromConnectUser,
		sessionChecker:        connectSessionChecker,
		serverErrorHandleFunc: defaultErrorHandler,
	}
}
