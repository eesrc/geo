package auth

import (
	"database/sql/driver"

	"github.com/eesrc/geo/pkg/serializing"
)

// Profile is the aggregated user which can be created from several different IDs. This
// in turn means that the Profile have several optional values and needs to be checked
// if the value is important to your user. The LoginID and Provider is the only fields that are
// mandatory, but the LoginID is only unique within the Provider.
type Profile struct {
	// Name of the user
	Name string `json:"name"`
	// PhoneNumber is the phone number of the user
	PhoneNumber string `json:"phone_number"`
	// PhoneNumberVerified tells if the phone number has been verified
	PhoneNumberVerified bool `json:"phone_number_verified"`
	// Email is the email of the user
	Email string `json:"email"`
	// VerfiedEmail tells if the email has been verified
	EmailVerified bool `json:"email_verified"`
	// LoginID is the ID used to uniquely identify the user
	LoginID string `json:"sub"`
	// AvatarURL is the URL to the user avatar
	AvatarURL string `json:"avatar_url"`
	// Provider determines what kind of ID the user is registered with
	Provider AuthProvider `json:"provider"`
	// UserObject is the raw user object retrieved from the Oauth Server
	UserObject interface{} `json:"user"`
}

// Scan implements the sql.Scanner interface
func (p *Profile) Scan(src interface{}) error {
	return serializing.ScanJSON(p, src)
}

// Value implements the driver.Valuer interface
func (p *Profile) Value() (driver.Value, error) {
	return serializing.ValueJSON(p)
}
