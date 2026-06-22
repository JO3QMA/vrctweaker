package activity

import "time"

// ActivityCommand is a fine-grained persistence intent produced by SessionCorrelator.
type ActivityCommand interface {
	activityCommand()
}

// EndPlaySessionCmd closes the latest open play session.
type EndPlaySessionCmd struct {
	At time.Time
}

func (EndPlaySessionCmd) activityCommand() {}

// StartPlaySessionCmd opens a new play session.
type StartPlaySessionCmd struct {
	InstanceID string
	At         time.Time
}

func (StartPlaySessionCmd) activityCommand() {}

// CloseOpenEncountersAtCmd sets left_at on all open encounter rows.
type CloseOpenEncountersAtCmd struct {
	At time.Time
}

func (CloseOpenEncountersAtCmd) activityCommand() {}

// RecordEncounterJoinCmd opens a new user encounter stay.
type RecordEncounterJoinCmd struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string
	At          time.Time
}

func (RecordEncounterJoinCmd) activityCommand() {}

// RecordEncounterLeaveCmd closes the user's open encounter stay.
type RecordEncounterLeaveCmd struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string
	At          time.Time
}

func (RecordEncounterLeaveCmd) activityCommand() {}

// UpsertWorldVisitCmd records a world visit from Destination set lines.
type UpsertWorldVisitCmd struct {
	WorldID string
	At      time.Time
}

func (UpsertWorldVisitCmd) activityCommand() {}

// UpsertWorldRoomNameCmd sets world display name from Entering Room lines.
type UpsertWorldRoomNameCmd struct {
	WorldID  string
	RoomName string
	At       time.Time
}

func (UpsertWorldRoomNameCmd) activityCommand() {}
