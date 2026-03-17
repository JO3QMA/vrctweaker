package activity

import "time"

// PlaySession represents a single VRChat play session.
type PlaySession struct {
	ID          string
	StartTime   time.Time
	EndTime     *time.Time
	DurationSec *int
}

// UserEncounter represents a join/leave event of a user in an instance.
type UserEncounter struct {
	ID            string
	VRCUserID     string
	DisplayName   string
	Action        string // "join" or "leave"
	InstanceID    string
	EncounteredAt time.Time
}
