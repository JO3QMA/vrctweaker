package activity

// SessionCorrelator maps parsed log events to fine-grained Activity commands using
// instance/world correlation state. It is pure (no I/O).
type SessionCorrelator struct {
	// session* are the active Joining instance (SessionEventStart only). Destination does not update these.
	sessionInstanceID string
	sessionWorldID    string
	// pendingDestinationWorldID is set by Destination set; survives SessionEventEnd for RoomName before Joining.
	pendingDestinationWorldID string
	// lastLeft* snapshot at SessionEventEnd for OnPlayerLeft lines that follow OnLeftRoom.
	lastLeftInstanceID string
	lastLeftWorldID    string
}

// Reset clears correlation state before reading a new output_log file from offset 0.
func (c *SessionCorrelator) Reset() {
	c.sessionInstanceID = ""
	c.sessionWorldID = ""
	c.pendingDestinationWorldID = ""
	c.lastLeftInstanceID = ""
	c.lastLeftWorldID = ""
}

// Apply consumes one parsed event and returns commands to persist. AvatarSwitch and VideoPlayback
// yield nil (out of scope until a consumer exists).
func (c *SessionCorrelator) Apply(event ParsedEvent) []ActivityCommand {
	if event == nil {
		return nil
	}
	switch e := event.(type) {
	case *DestinationSetEvent:
		c.pendingDestinationWorldID = e.WorldID
		return []ActivityCommand{UpsertWorldVisitCmd{WorldID: e.WorldID, At: e.OccurredAt}}
	case *RoomNameEvent:
		wid := c.sessionWorldID
		if wid == "" {
			wid = c.pendingDestinationWorldID
		}
		return []ActivityCommand{UpsertWorldRoomNameCmd{
			WorldID:  wid,
			RoomName: e.RoomName,
			At:       e.OccurredAt,
		}}
	case *EncounterEvent:
		inst := e.InstanceID
		if inst == "" {
			inst = c.sessionInstanceID
		}
		if inst == "" && e.Action == EncounterActionLeave {
			inst = c.lastLeftInstanceID
		}
		wid := c.sessionWorldID
		if wid == "" && e.Action == EncounterActionLeave {
			wid = c.lastLeftWorldID
		}
		if wid == "" {
			wid = c.pendingDestinationWorldID
		}
		if e.Action == EncounterActionJoin {
			return []ActivityCommand{RecordEncounterJoinCmd{
				VRCUserID:   e.VRCUserID,
				DisplayName: e.DisplayName,
				InstanceID:  inst,
				WorldID:     wid,
				At:          e.EncounteredAt,
			}}
		}
		return []ActivityCommand{RecordEncounterLeaveCmd{
			VRCUserID:   e.VRCUserID,
			DisplayName: e.DisplayName,
			InstanceID:  inst,
			WorldID:     wid,
			At:          e.EncounteredAt,
		}}
	case *SessionEvent:
		return c.applySession(e)
	default:
		return nil
	}
}

func (c *SessionCorrelator) applySession(e *SessionEvent) []ActivityCommand {
	switch e.Type {
	case SessionEventStart:
		if e.InstanceID == "" {
			return nil
		}
		c.lastLeftInstanceID = ""
		c.lastLeftWorldID = ""
		c.sessionInstanceID = e.InstanceID
		if w := WorldIDFromInstanceKey(e.InstanceID); w != "" {
			c.sessionWorldID = w
		} else {
			c.sessionWorldID = ""
		}
		c.pendingDestinationWorldID = ""
		return []ActivityCommand{
			EndPlaySessionCmd{At: e.OccurredAt},
			CloseOpenEncountersAtCmd{At: e.OccurredAt},
			StartPlaySessionCmd{InstanceID: e.InstanceID, At: e.OccurredAt},
		}
	case SessionEventEnd:
		c.lastLeftInstanceID = c.sessionInstanceID
		c.lastLeftWorldID = c.sessionWorldID
		c.sessionInstanceID = ""
		c.sessionWorldID = ""
		return []ActivityCommand{EndPlaySessionCmd{At: e.OccurredAt}}
	default:
		return nil
	}
}
