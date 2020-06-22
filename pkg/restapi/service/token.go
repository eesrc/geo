package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eesrc/geo/pkg/model"
)

// Token is the API representation of a token
type Token struct {
	Token     string    `json:"token"`
	Created   time.Time `json:"created"`
	Resource  string    `json:"resource"`
	PermWrite bool      `json:"permWrite"`
	UserID    int64     `json:"-"`
}

// ToModel creates a storage model from the API representation
func (token *Token) ToModel() *model.Token {
	return &model.Token{
		Token:     token.Token,
		Created:   token.Created,
		Resource:  token.Resource,
		PermWrite: token.PermWrite,
		UserID:    token.UserID,
	}
}

// GenerateToken overwrites the current token value with a random generated hex string
func (t *Token) GenerateToken() error {
	buf := make([]byte, 32)
	n, err := rand.Read(buf)
	if err == nil && n != len(buf) {
		return fmt.Errorf("unable to generate token %d bytes long. Only got %d bytes", len(buf), n)
	}
	t.Token = hex.EncodeToString(buf)
	return nil
}

// MarshalJSON marshals a JSON string from the API representation
func (token *Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(*token)
}

// NewTokenFromModel creates a HTTP representation of a model token
func NewTokenFromModel(tokenModel *model.Token) *Token {
	return &Token{
		Token:     tokenModel.Token,
		Created:   tokenModel.Created,
		Resource:  tokenModel.Resource,
		PermWrite: tokenModel.PermWrite,
		UserID:    tokenModel.UserID,
	}
}
