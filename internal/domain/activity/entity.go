package activity

import "time"

// PlaySession represents a single VRChat play session.
type PlaySession struct {
	ID          string
	StartTime   time.Time
	EndTime     *time.Time
	DurationSec *int
}

// UserEncounter represents one stay (join → leave) of a user in an instance.
// LeftAt nil means the stay is still open (no leave observed yet).
type UserEncounter struct {
	ID          string
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string // wrld_* from current instance when known
	JoinedAt    time.Time
	LeftAt      *time.Time
}

// EncounterWithContext is a user encounter plus joined user/world cache fields for the UI.
type EncounterWithContext struct {
	Encounter         *UserEncounter
	WorldDisplayName  string
	UserFirstSeenAt   *time.Time
	UserLastContactAt *time.Time
	IsFirstEncounter  bool
}
