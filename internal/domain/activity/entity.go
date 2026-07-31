package activity

import "time"

// PlaySession represents a single VRChat play session.
type PlaySession struct {
	ID            string
	StartTime     time.Time
	EndTime       *time.Time
	DurationSec   *int
	InstanceID    string // VRChat instance key when known
	LogSourcePath string // output_log absolute path; empty for legacy rows
}

// UserEncounter represents one stay (join → leave) of a user in an instance.
// LeftAt nil means the stay is still open (no leave observed yet).
type UserEncounter struct {
	ID            string
	VRCUserID     string
	DisplayName   string
	InstanceID    string
	WorldID       string // wrld_* from current instance when known
	LogSourcePath string // output_log absolute path; empty for legacy rows
	JoinedAt      time.Time
	LeftAt        *time.Time
}

// EncounterWithContext is a user encounter plus joined user/world cache fields for the UI.
type EncounterWithContext struct {
	Encounter         *UserEncounter
	WorldDisplayName  string
	UserFirstSeenAt   *time.Time
	UserLastContactAt *time.Time
	IsFirstEncounter  bool
	// IsListableFriend is true when users_cache marks the user as a Listable friend at list time.
	IsListableFriend bool
}

// EncounterFilter provides optional filtering for encounter list queries.
type EncounterFilter struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string
	From        *time.Time
	To          *time.Time
}

// Video playback outcome values persisted on VideoPlaybackAttempt.Outcome.
const (
	VideoPlaybackOutcomeOpen    = ""
	VideoPlaybackOutcomeSuccess = "success"
	VideoPlaybackOutcomeFailure = "failure"
)

// VideoPlaybackAttempt is one URL resolve try from [Video Playback] logs.
type VideoPlaybackAttempt struct {
	ID            string
	AttemptedAt   time.Time
	URL           string
	Outcome       string // VideoPlaybackOutcome*
	FailureReason string
	ResolvedURL   string
	WorldID       string
	LogSourcePath string
	CompletedAt   *time.Time
}

// VideoPlaybackWithContext is a video playback attempt plus world display name for the UI.
type VideoPlaybackWithContext struct {
	Attempt          *VideoPlaybackAttempt
	WorldDisplayName string
}
