package auth

// Session is the stored session data
type Session struct {
	ID          string
	Expires     int64
	AccessToken string
	Profile     Profile
}

// SessionStore is the interface for the session back-end store.
type SessionStore interface {
	// PutState inserts a new state nonce in the storage. The nonce will never expire
	PutState(state string) error

	// RemoveState removes a state nonce from the storage. An error is returned if the
	// nonce does not exist.
	RemoveState(state string) error

	// CreateSession creates a new session in the store. The expires parameter is the expire
	// time in nanonseconds for the session.
	CreateSession(accessToken string, expires int64, profile Profile) (string, error)

	// GetSession returns the session from the session store based on given ID and where expire is higher than provided nanoseconds.
	GetSession(sessionID string, ingnoreOlderNs int64) (Session, error)

	// RemoveSession removes the session from the store.
	RemoveSession(sessionID string) error

	// GetSessions returns sessions with expire time is less than given time in nanoseconds.
	GetSessions(time int64) ([]Session, error)

	// RefreshSession refreshes a session expire time by adding the given nanoseconds.
	RefreshSession(sessionID string, expireAddNs int64) error
}
