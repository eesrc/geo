package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/restapi/input"
	"github.com/eesrc/geo/pkg/restapi/service"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/store/errors"
)

// GetTokenID returns a tokenID from given HandlerParameterMap. If missing or corrupt, returns a
// validationError
func GetTokenID(handlerParams HandlerParameterMap) (string, error) {
	tokenID, err := handlerParams.TokenID()
	if err != nil {
		return "", newError(NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail("tokenId", fmt.Sprintf("The token id '%s' is malformed", handlerParams["tokenID"])),
		))
	}

	return tokenID, nil
}

// GetTokenFromHandlerParams validates token (if) found in params and returns a token from store.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetTokenFromHandlerParams(handlerParams HandlerParameterMap, userID int64, store store.Store) (*service.Token, error) {
	tokenID, err := GetTokenID(handlerParams)
	if err != nil {
		return &service.Token{}, err
	}

	token, err := GetToken(tokenID, userID, store)
	if err != nil {
		return &service.Token{}, err
	}

	return token, nil
}

// GetToken returns a token from store. Returns a validation error or regular error
// if the fetch fails.
func GetToken(tokenID string, userId int64, store store.Store) (*service.Token, error) {
	token, err := store.GetTokenByUserID(tokenID, userId)

	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			if storageError.Type == errors.AccessDeniedError || storageError.Type == errors.NotFoundError {
				return &service.Token{}, newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("tokenId", fmt.Sprintf("The token with id '%s' might not exist", tokenID)),
				))
			}
		}

		return &service.Token{}, err
	}

	return service.NewTokenFromModel(token), nil
}

// CreateToken tries to create a token and returns a regular error
// if the creation fails
func CreateToken(collection *model.Token, store store.Store) (string, error) {
	return store.CreateToken(collection)
}

// UpdateToken tries to update a token and returns a validation error or regular error
// if the update fails
func UpdateToken(token *model.Token, store store.Store) error {
	err := store.UpdateToken(token)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("token.token", fmt.Sprintf("The token id '%s' might not exist", token.Token)),
				))
			}
		}
	}

	return err
}

// DeleteToken tries to delete a token and returns a validation error or regular error
// if the delete fails
func DeleteToken(tokenID string, userID int64, store store.Store) error {
	err := store.DeleteToken(tokenID, userID)

	// Check if there's a reason to create a validationError
	if err != nil {
		if storageError, ok := err.(*errors.StorageError); ok {
			switch storageError.Type {
			// AccessDenied and NotFound are handled the same
			case errors.AccessDeniedError, errors.NotFoundError:
				return newError(NewErrorResponse(
					http.StatusNotFound,
					NewParameterErrorDetail("tokenId", fmt.Sprintf("The token with id '%s' might not exist", tokenID)),
				))
			}
		}
	}

	return err
}

// ListTokens lists tokens based on userID and returns a store error if the list fails
func ListTokens(userID int64, filterParams FilterParams, store store.Store) ([]*service.Token, error) {
	tokens, err := store.ListTokensByUserID(userID, filterParams.Offset, filterParams.Limit)
	if err != nil {
		return []*service.Token{}, err
	}

	var tokenList []*service.Token = make([]*service.Token, len(tokens))

	for i, token := range tokens {
		tokenList[i] = service.NewTokenFromModel(&token)
	}

	return tokenList, nil
}

// GetTokenFromBody retrieves a token from given body and decodes it.
// Returns a validation error containing an ErrorResponse if something went wrong
func GetTokenFromBody(body io.ReadCloser) (*service.Token, error) {
	jsonDecoder := json.NewDecoder(body)
	var token service.Token
	err := jsonDecoder.Decode(&token)
	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &token, getUnmarshalError(err)
		}

		return &token, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("token", "You need to provide a valid token object"),
			),
		)
	}

	// Clean input
	token.Resource = input.NormalizeStringInput(token.Resource)

	return &token, nil
}
