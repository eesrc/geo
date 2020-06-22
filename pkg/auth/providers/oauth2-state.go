package providers

import (
	"crypto/rand"
	"encoding/hex"
)

// makeState creates a state token for the OAuth service. It's just a random string of bytes
func newState() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
