package identity

import "time"

// UserCache represents a VRChat user row in users_cache (self, friends, and log-derived contacts).
type UserCache struct {
	VRCUserID     string
	DisplayName   string
	Status        string // join me, active, offline, etc.; empty when only seen in logs
	IsFavorite    bool
	LastUpdated   time.Time
	FirstSeenAt   *time.Time
	LastContactAt *time.Time
	UserKind      UserKind
	// SessionFingerprint scopes the self row to the current auth token (hex SHA-256).
	SessionFingerprint string
	// Self-profile fields (GET /auth/user); also stored for user_kind=self rows.
	Username                    string
	StatusDescription           string
	UserState                   string
	AvatarThumbnailURL          string
	UserIconURL                 string
	ProfilePicOverrideThumbnail string
}
