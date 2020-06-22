package providers

import (
	"encoding/json"

	"github.com/eesrc/geo/pkg/auth"
)

// ConnectUserProfile contains the Connect profile.
type ConnectUserProfile struct {
	ID                  string `json:"sub"`
	Name                string `json:"name"`
	Email               string `json:"email"`
	EmailVerified       bool   `json:"email_verified"`
	PhoneNumber         string `json:"phone_number"`
	PhoneVerifiedNumber bool   `json:"phone_number_verified"`
}

// profileFromConnectUser reads the profile from a the dictionary returned by the
// Connect User API.
func profileFromConnectUser(jsonBytes []byte) auth.Profile {
	var connectProfile ConnectUserProfile
	profile := auth.Profile{}

	err := json.Unmarshal(jsonBytes, &connectProfile)

	if err != nil {
		return profile
	}

	profile.LoginID = connectProfile.ID

	profile.Name = connectProfile.Name

	profile.Email = connectProfile.Email
	profile.EmailVerified = connectProfile.EmailVerified

	profile.UserObject = connectProfile
	profile.Provider = auth.AuthConnect

	return profile
}
