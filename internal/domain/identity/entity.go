package identity

import "time"

// UserCache represents a VRChat user row in users_cache (API friends and/or log-derived contacts).
type UserCache struct {
	VRCUserID     string
	DisplayName   string
	Status        string // join me, active, offline, etc.; empty when only seen in logs
	IsFavorite    bool
	LastUpdated   time.Time
	FirstSeenAt   *time.Time
	LastContactAt *time.Time
}
