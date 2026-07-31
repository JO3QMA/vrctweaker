package activity

import "time"

// EndPlaySessionCmd closes the latest open play session.
type EndPlaySessionCmd struct {
	At time.Time
}

// StartPlaySessionCmd opens a new play session.
type StartPlaySessionCmd struct {
	InstanceID string
	At         time.Time
}

// CloseOpenEncountersAtCmd sets left_at on all open encounter rows.
type CloseOpenEncountersAtCmd struct {
	At time.Time
}

// RecordEncounterJoinCmd opens a new user encounter stay.
type RecordEncounterJoinCmd struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string
	At          time.Time
}

// RecordEncounterLeaveCmd closes the user's open encounter stay.
type RecordEncounterLeaveCmd struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string
	At          time.Time
}

// UpsertWorldVisitCmd records a world visit from Destination set lines.
type UpsertWorldVisitCmd struct {
	WorldID string
	At      time.Time
}

// UpsertWorldRoomNameCmd sets world display name from Entering Room lines.
type UpsertWorldRoomNameCmd struct {
	WorldID  string
	RoomName string
	At       time.Time
}

// RecordVideoPlaybackAttemptCmd opens a Video playback attempt row.
type RecordVideoPlaybackAttemptCmd struct {
	URL     string
	WorldID string // sessionWorldID at attempt time; empty if no Open play session
	At      time.Time
}

// CompleteVideoPlaybackFailureCmd marks the matching Open attempt as failure.
// URL empty means match the oldest Open on this Log source (ERROR lines often omit URL).
type CompleteVideoPlaybackFailureCmd struct {
	URL           string
	FailureReason string
	At            time.Time
}

// CompleteVideoPlaybackSuccessCmd marks the oldest Open attempt with URL as success,
// only if still open (ERROR-first: already-failed rows are not overwritten).
type CompleteVideoPlaybackSuccessCmd struct {
	URL         string
	ResolvedURL string
	At          time.Time
}
