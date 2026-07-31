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
	// pendingVideo is FIFO of Open video playback attempt URLs on this Log source.
	pendingVideo []string
}

// Reset clears correlation state before reading a new output_log file from offset 0.
func (c *SessionCorrelator) Reset() {
	c.sessionInstanceID = ""
	c.sessionWorldID = ""
	c.pendingDestinationWorldID = ""
	c.lastLeftInstanceID = ""
	c.lastLeftWorldID = ""
	c.pendingVideo = nil
}

// Apply consumes one parsed event and returns commands to persist. AvatarSwitch yields nil.
func (c *SessionCorrelator) Apply(event ParsedEvent) []any {
	if event == nil {
		return nil
	}
	switch e := event.(type) {
	case *DestinationSetEvent:
		c.pendingDestinationWorldID = e.WorldID
		return []any{UpsertWorldVisitCmd{WorldID: e.WorldID, At: e.OccurredAt}}
	case *RoomNameEvent:
		// Entering Room names the destination world. During a transition the old session may still
		// be active until OnLeftRoom; prefer pending Destination set over sessionWorldID.
		wid := c.pendingDestinationWorldID
		if wid == "" {
			wid = c.sessionWorldID
		}
		return []any{UpsertWorldRoomNameCmd{
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
			return []any{RecordEncounterJoinCmd{
				VRCUserID:   e.VRCUserID,
				DisplayName: e.DisplayName,
				InstanceID:  inst,
				WorldID:     wid,
				At:          e.EncounteredAt,
			}}
		}
		return []any{RecordEncounterLeaveCmd{
			VRCUserID:   e.VRCUserID,
			DisplayName: e.DisplayName,
			InstanceID:  inst,
			WorldID:     wid,
			At:          e.EncounteredAt,
		}}
	case *SessionEvent:
		return c.applySession(e)
	case *VideoPlaybackEvent:
		return c.applyVideoAttempt(e)
	case *VideoPlaybackErrorEvent:
		return c.applyVideoError(e)
	case *VideoPlaybackResolvedEvent:
		return c.applyVideoResolved(e)
	default:
		return nil
	}
}

func (c *SessionCorrelator) applyVideoAttempt(e *VideoPlaybackEvent) []any {
	if e.URL == "" {
		return nil
	}
	c.pendingVideo = append(c.pendingVideo, e.URL)
	return []any{RecordVideoPlaybackAttemptCmd{
		URL:     e.URL,
		WorldID: c.sessionWorldID, // Open play session only; do not use lastLeft
		At:      e.OccurredAt,
	}}
}

func (c *SessionCorrelator) applyVideoError(e *VideoPlaybackErrorEvent) []any {
	idx := c.findPendingVideoIndex(e.URL)
	if idx < 0 {
		return nil // orphan result
	}
	url := c.pendingVideo[idx]
	c.pendingVideo = append(c.pendingVideo[:idx], c.pendingVideo[idx+1:]...)
	return []any{CompleteVideoPlaybackFailureCmd{
		URL:           url,
		FailureReason: e.Message,
		At:            e.OccurredAt,
	}}
}

func (c *SessionCorrelator) applyVideoResolved(e *VideoPlaybackResolvedEvent) []any {
	if e.URL == "" {
		return nil
	}
	idx := c.findPendingVideoIndex(e.URL)
	if idx < 0 {
		return nil // orphan, or already failed (removed from pending on ERROR)
	}
	c.pendingVideo = append(c.pendingVideo[:idx], c.pendingVideo[idx+1:]...)
	return []any{CompleteVideoPlaybackSuccessCmd{
		URL:         e.URL,
		ResolvedURL: e.ResolvedURL,
		At:          e.OccurredAt,
	}}
}

// findPendingVideoIndex returns the oldest pending index matching url.
// When url is empty (typical ERROR line), returns the oldest Open overall.
func (c *SessionCorrelator) findPendingVideoIndex(url string) int {
	if len(c.pendingVideo) == 0 {
		return -1
	}
	if url == "" {
		return 0
	}
	for i, u := range c.pendingVideo {
		if u == url {
			return i
		}
	}
	return -1
}

func (c *SessionCorrelator) applySession(e *SessionEvent) []any {
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
		return []any{
			EndPlaySessionCmd{At: e.OccurredAt},
			CloseOpenEncountersAtCmd{At: e.OccurredAt},
			StartPlaySessionCmd{InstanceID: e.InstanceID, At: e.OccurredAt},
		}
	case SessionEventEnd:
		c.lastLeftInstanceID = c.sessionInstanceID
		c.lastLeftWorldID = c.sessionWorldID
		c.sessionInstanceID = ""
		c.sessionWorldID = ""
		return []any{EndPlaySessionCmd{At: e.OccurredAt}}
	default:
		return nil
	}
}
