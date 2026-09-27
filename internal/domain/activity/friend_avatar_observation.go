package activity

import "time"

// FriendAvatarObservation is one log-derived avatar switch for a user.
type FriendAvatarObservation struct {
	ID            string
	VRCUserID     string
	DisplayName   string
	AvatarName    string
	InstanceID    string
	WorldID       string
	LogSourcePath string
	ObservedAt    time.Time
}

// FriendAvatarUsageSummary aggregates observations per avatar name for one friend.
type FriendAvatarUsageSummary struct {
	AvatarName  string
	UseCount    int64
	FirstSeenAt time.Time
	LastSeenAt  time.Time
}
