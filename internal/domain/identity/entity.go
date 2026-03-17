package identity

import "time"

// FriendCache represents cached friend information from VRChat API.
type FriendCache struct {
	VRCUserID   string
	DisplayName string
	Status      string // join me, active, offline, etc.
	IsFavorite  bool
	LastUpdated time.Time
}
