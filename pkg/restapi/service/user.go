package service

import (
	"encoding/json"
	"time"

	"github.com/eesrc/geo/pkg/model"
)

// User is the API representation of a user
type User struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Phone         string    `json:"phone"`
	PhoneVerified bool      `json:"phoneVerified"`
	Created       time.Time `json:"created"`
}

// MarshalJSON marshals a JSON string from the API representation
func (user *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(*user)
}

// NewUserFromModel creates a HTTP representation of a model user
func NewUserFromModel(userModel *model.User) *User {
	return &User{
		ID:            userModel.ID,
		Name:          userModel.Name,
		Email:         userModel.Email,
		EmailVerified: userModel.EmailVerified,
		Phone:         userModel.Phone,
		PhoneVerified: userModel.PhoneVerified,
		Created:       userModel.Created,
	}
}
