package identity

import (
	"crypto/sha256"
	"encoding/hex"
)

// AuthTokenFingerprint returns a stable hash of the auth token for cache scoping (not for security).
func AuthTokenFingerprint(token string) string {
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
