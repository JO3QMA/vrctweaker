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
