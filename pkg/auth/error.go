package auth

import (
	"errors"
)

// AuthError is an error specific to an error in the OAuth2 or session process
type AuthError error

var (
	// Errors received from OAuth server, ofte connected to wrongly configured OAuth client

	RedirectURIMismatchError        AuthError = errors.New("The redirect URL provided is a mismatch to the configured OAuth application")
	IncorrectClientCredentialsError AuthError = errors.New("The provided client id or client secret are incorrect")
	BadVerificationCode             AuthError = errors.New("The code passed is either incorrect or expired")
	InvalidScopeError               AuthError = errors.New("The provided scope is not valid for the provider")
	UserAccessDeniedError           AuthError = errors.New("The user denied the authorization request")
	ServerError                     AuthError = errors.New("The provider returned server error")
	TemporarilyUnavailableError     AuthError = errors.New("The provider is temporarily unavailable")

	// Local errors

	MissingCodeError  AuthError = errors.New("No access token code in query")
	UnknownStateError AuthError = errors.New("Unknown state variable in query")
	RemoteServerError AuthError = errors.New("Something went wrong during requesting a remote server")

	UnknownMethodError AuthError = errors.New("Method or path not supported")
	PersistStateError  AuthError = errors.New("Failed to persist state in Session")
	StoreError         AuthError = errors.New("Failed to persist profile in DB")

	UnauthorizedError AuthError = errors.New("User is not authorized")
)
