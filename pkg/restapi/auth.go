package restapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/auth"
	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/store/errors"
	"github.com/gorilla/mux"
)

func (s *Server) addAuthHandlers(router *mux.Router) {
	// Set default handlers for authenticator in case of 5xx
	s.authenticator.ServerErrorHandlerFunc(serverErrorHandlerFunc)

	for _, provider := range s.authenticator.Providers {
		provider.StartSessionChecker()
		router.PathPrefix(provider.BasePath()).Handler(provider)
	}
}

// Keys for request contexts
type contextKey string

// The key used in the request context
const userKey = contextKey("user")
const authTypeKey = contextKey("auth")

// authSessionToUserHandlerFunc grabs the session from the context and
// injects the actual user information into the context. If it fails to fetch the user,
// it will return 503 for the user.
func (s *Server) authSessionToUserHandlerFunc(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newContext := r.Context()

		// Check token authentication
		if tokenString := r.Header.Get("X-API-token"); tokenString != "" {
			token, err := s.store.GetToken(tokenString)

			if err != nil {
				log.Warnf("Error when trying to fetch token: %v", err)
				errorResponse := validation.NewErrorResponse(http.StatusUnauthorized)
				errorResponse.WriteHTTPError(w)
				return
			}

			if !token.PermWrite && (r.Method != http.MethodGet && r.Method != http.MethodOptions) {
				// User is trying to access methods which requires permission write
				validation.NewErrorResponse(
					http.StatusForbidden,
					validation.NewParameterErrorDetail(
						"permWrite",
						"The token used does not have write permission and can only use the HTTP verbs GET and OPTIONS",
					),
				).WriteHTTPError(w)
				return
			}

			if token.Resource != "" && !strings.HasPrefix(r.RequestURI, token.Resource) {
				// Token is restricted by the resource path
				validation.NewErrorResponse(
					http.StatusForbidden,
					validation.NewParameterErrorDetail(
						"resource",
						fmt.Sprintf("The token only have access to the resource '%s'", token.Resource),
					),
				).WriteHTTPError(w)
				return
			}

			user, err := s.store.GetUser(token.UserID)
			if err != nil {
				// Dangling token, ie the user does not exists anymore
				log.Errorf("Error when trying to fetch user for token: %v", err)
				errorResponse := validation.NewErrorResponse(http.StatusUnauthorized)
				errorResponse.WriteHTTPError(w)
				return
			}

			newContext = context.WithValue(newContext, userKey, user)
			newContext = context.WithValue(newContext, authTypeKey, auth.AuthToken)

			f.ServeHTTP(w, r.WithContext(newContext))
			return
		}

		// Check authenticator and providers
		if s.authenticator == nil {
			log.Errorf("Authenticator not set")
			validation.NewErrorResponse(http.StatusInternalServerError).WriteHTTPError(w)
			return
		}

		profile, err := s.authenticator.Profile(w, r)

		if err != nil {
			validation.NewErrorResponse(http.StatusUnauthorized).WriteHTTPError(w)
			return
		}

		switch profile.Provider {
		case auth.AuthGithub:
			user, err := s.addOrUpdateUser(profile)
			if err != nil {
				log.Errorf("Could not add or update user: %v", err)
				validation.NewErrorResponse(http.StatusServiceUnavailable).WriteHTTPError(w)
				return
			}
			newContext = context.WithValue(newContext, userKey, user)
			newContext = context.WithValue(newContext, authTypeKey, auth.AuthGithub)
		case auth.AuthConnect:
			user, err := s.addOrUpdateUser(profile)
			if err != nil {
				log.Errorf("Could not add or update user: %v", err)
				validation.NewErrorResponse(http.StatusServiceUnavailable).WriteHTTPError(w)
				return
			}
			newContext = context.WithValue(newContext, userKey, user)
			newContext = context.WithValue(newContext, authTypeKey, auth.AuthConnect)
		default:
			log.Errorf("Error when trying to retrieve profile. Unknown provider %s", profile.Provider)
		}

		f.ServeHTTP(w, r.WithContext(newContext))
	})
}

// addOrUpdateUser takes a profile session and does one of the following:
// 1. If the user does not exist, it creates a new user based on the profile along
// with a private team and an initial collection. It then returns the newly created user.
//
// 2. If the user exists, it checks towards the profile if there's any updates to be made
// to the persisted user. If so, it updates the persisted user and returns the updated user.
//
// 3. The user does not exist and an error occured when trying to add it. Returns nil and error.
func (s *Server) addOrUpdateUser(profile auth.Profile) (*model.User, error) {
	var user *model.User
	var err error

	switch profile.Provider {
	case auth.AuthGithub:
		user, err = s.store.GetUserByGithubID(profile.LoginID)
	case auth.AuthConnect:
		user, err = s.store.GetUserByConnectID(profile.LoginID)
	default:
		return nil, fmt.Errorf("Failed to get user by given provider %v", profile.Provider)
	}

	if storageError, ok := err.(*errors.StorageError); ok {
		if storageError.Type == errors.NotFoundError {
			log.Info("User not found in DB, creating user")
			return s.provisionNewUser(profile)
		}
	}

	if user != nil {
		// Check on diff of provided profile and user for attributes we're interested in having updated.
		if user.Name != profile.Name ||
			user.Email != profile.Email {
			log.Info("User found in DB, but the user has changed. Updating with new name and email")

			user.Name = profile.Name
			user.Email = profile.Email

			if err := s.store.UpdateUser(user); err != nil {
				log.WithError(err).Errorf("Unable to update user %v", user)
			}
		}
	}

	return user, nil
}

// UserFromRequest returns the User model of the currenly logged in user.
func (s *Server) UserFromRequest(r *http.Request) *model.User {
	if r == nil || r.Context() == nil {
		return nil
	}
	v := r.Context().Value(userKey)
	if v == nil {
		return nil
	}
	user, ok := v.(*model.User)
	if !ok {
		return nil
	}
	return user
}

// AuthTypeKeyFromRequest returns the user auth type
func (s *Server) AuthTypeKeyFromRequest(r *http.Request) auth.AuthProvider {
	if r == nil || r.Context() == nil {
		return ""
	}
	v := r.Context().Value(authTypeKey)
	if v == nil {
		return ""
	}
	userAuthType, ok := v.(auth.AuthProvider)
	if !ok {
		return ""
	}
	return userAuthType
}

// provisionNewUser provisions a new User based on a Profile
func (s *Server) provisionNewUser(profile auth.Profile) (*model.User, error) {
	var user *model.User

	switch profile.Provider {
	case auth.AuthGithub:
		user = &model.User{
			Email:         profile.Email,
			EmailVerified: true,
			Name:          profile.Name,
			Phone:         "",
			PhoneVerified: false,
			GithubID:      profile.LoginID,
			Created:       time.Now(),
		}
	case auth.AuthConnect:
		user = &model.User{
			Email:         profile.Email,
			EmailVerified: true,
			Name:          profile.Name,
			Phone:         "",
			PhoneVerified: false,
			ConnectID:     profile.LoginID,
			Created:       time.Now(),
		}
	default:
		return nil, fmt.Errorf("Can't create user with provider, %s", profile.Provider)
	}

	userID, err := s.store.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("Failed to create user, %v", err)
	}

	// Create an initial Team for the new User
	teamID, err := s.store.CreateTeam(&model.Team{
		Name: "My private team",
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to create private team, %v", err)
	}

	err = s.store.SetTeamMember(userID, teamID, true)

	if err != nil {
		return nil, fmt.Errorf("Failed to set team member, %v", err)
	}

	// Create an initial Collection for the new user
	_, err = s.store.CreateCollection(&model.Collection{
		TeamID: teamID,
		Name:   "My default collection",
	}, userID)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new default collection, %v", err)
	}

	user, err = s.store.GetUser(userID)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch new user after create, %v", err)
	}

	return user, nil
}

func serverErrorHandlerFunc(w http.ResponseWriter, r *http.Request, err auth.AuthError) {
	log.Error("Got an server error: ", err)
	errorResponse := validation.NewErrorResponse(http.StatusServiceUnavailable)
	errorResponse.WriteHTTPError(w)
}
