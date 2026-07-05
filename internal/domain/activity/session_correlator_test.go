package activity

import (
	"testing"
	"time"

	"vrchat-tweaker/internal/testvrc"
)

const testWorldID = "wrld_beddab1e-fee1-cafe-f00d-ca7c0dd1eca7"

var testFullInstance = testWorldID + ":41550~hidden(" + testvrc.EmbedUserID + ")~region(jp)"

func TestSessionCorrelator_RoomNameAfterOnLeftRoom_usesPendingDestinationWorld(t *testing.T) {
	base := time.Date(2026, 3, 22, 11, 23, 51, 0, time.UTC)
	c := &SessionCorrelator{}

	c.Apply(&DestinationSetEvent{
		WorldID:      testWorldID,
		FullInstance: testFullInstance,
		OccurredAt:   base,
	})
	c.Apply(&SessionEvent{Type: SessionEventEnd, OccurredAt: base})
	cmds := c.Apply(&RoomNameEvent{RoomName: "SuRroom", OccurredAt: base})
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: testFullInstance, OccurredAt: base})

	if len(cmds) != 1 {
		t.Fatalf("RoomName commands = %d, want 1: %+v", len(cmds), cmds)
	}
	room, ok := cmds[0].(UpsertWorldRoomNameCmd)
	if !ok {
		t.Fatalf("command type = %T, want UpsertWorldRoomNameCmd", cmds[0])
	}
	if room.WorldID != testWorldID || room.RoomName != "SuRroom" {
		t.Errorf("UpsertWorldRoomName = %+v, want world %q name %q", room, testWorldID, "SuRroom")
	}
}

func TestSessionCorrelator_ResetBetweenLogFiles_RoomNameUsesPendingDestination(t *testing.T) {
	base := time.Date(2026, 3, 18, 0, 30, 0, 0, time.UTC)
	const prevWorld = "wrld_c03f8195-3c64-46d8-b5ae-242f214c9404"
	prevInst := prevWorld + ":98225~hidden(" + testvrc.HiddenHostUserID + ")~region(jp)"
	const nextWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
	nextInst := nextWorld + ":77788~private(" + testvrc.PlayerUserID + ")~region(jp)"

	c := &SessionCorrelator{}
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: prevInst, OccurredAt: base})
	c.Reset()
	c.Apply(&DestinationSetEvent{
		WorldID:      nextWorld,
		FullInstance: nextInst,
		OccurredAt:   base,
	})
	cmds := c.Apply(&RoomNameEvent{RoomName: "ホームチェックv6․0", OccurredAt: base})

	if len(cmds) != 1 {
		t.Fatalf("commands = %d, want 1: %+v", len(cmds), cmds)
	}
	room := cmds[0].(UpsertWorldRoomNameCmd)
	if room.WorldID != nextWorld {
		t.Errorf("world_id = %q, want %q (must not use previous log file session)", room.WorldID, nextWorld)
	}
}

func TestSessionCorrelator_RoomNameWithoutOnLeftRoom_unchanged(t *testing.T) {
	base := time.Date(2026, 3, 22, 11, 22, 51, 0, time.UTC)
	const homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
	homeInst := homeWorld + ":04910~private(" + testvrc.PlayerUserID + ")~region(jp)"
	c := &SessionCorrelator{}

	c.Apply(&DestinationSetEvent{
		WorldID:      homeWorld,
		FullInstance: homeInst,
		OccurredAt:   base,
	})
	cmds := c.Apply(&RoomNameEvent{RoomName: "ホームチェックv6․0", OccurredAt: base})
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: homeInst, OccurredAt: base})

	room := cmds[0].(UpsertWorldRoomNameCmd)
	if room.WorldID != homeWorld || room.RoomName != "ホームチェックv6․0" {
		t.Errorf("UpsertWorldRoomName = %+v", room)
	}
}

func TestSessionCorrelator_RoomName_prefersPendingOverActiveSession(t *testing.T) {
	base := time.Date(2026, 6, 24, 8, 26, 44, 0, time.UTC)
	const oldWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
	oldInst := oldWorld + ":95147~private(" + testvrc.PlayerUserID + ")~region(jp)"
	const newWorld = "wrld_6041ba53-0ac0-4b5b-9ecb-890ea2b0aefa"
	newInst := newWorld + ":48580~friends(" + testvrc.FriendsHostUserID + ")~region(jp)"

	c := &SessionCorrelator{}
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: oldInst, OccurredAt: base})
	c.Apply(&DestinationSetEvent{
		WorldID:      newWorld,
		FullInstance: newInst,
		OccurredAt:   base,
	})
	cmds := c.Apply(&RoomNameEvent{RoomName: "Cozy with.", OccurredAt: base})

	if len(cmds) != 1 {
		t.Fatalf("commands = %d, want 1: %+v", len(cmds), cmds)
	}
	room := cmds[0].(UpsertWorldRoomNameCmd)
	if room.WorldID != newWorld || room.RoomName != "Cozy with." {
		t.Errorf("UpsertWorldRoomName = %+v, want world %q name %q", room, newWorld, "Cozy with.")
	}
}

func TestSessionCorrelator_OtherPlayerLeave_afterJoining_keepsWorldContext(t *testing.T) {
	base := time.Date(2026, 3, 18, 0, 1, 0, 0, time.UTC)
	const minasocoWorld = "wrld_c03f8195-3c64-46d8-b5ae-242f214c9404"
	minasocoInst := minasocoWorld + ":98225~hidden(" + testvrc.HiddenHostUserID + ")~region(jp)"
	otherUser := testvrc.OtherPlayerUserID

	c := &SessionCorrelator{}
	c.Apply(&DestinationSetEvent{
		WorldID:      minasocoWorld,
		FullInstance: minasocoInst,
		OccurredAt:   base,
	})
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: minasocoInst, OccurredAt: base})
	joinCmds := c.Apply(&EncounterEvent{
		VRCUserID:     otherUser,
		DisplayName:   "Nau_UoxoU",
		Action:        EncounterActionJoin,
		EncounteredAt: base,
	})
	leaveCmds := c.Apply(&EncounterEvent{
		VRCUserID:     otherUser,
		DisplayName:   "Nau_UoxoU",
		Action:        EncounterActionLeave,
		EncounteredAt: base.Add(time.Second),
	})

	join := joinCmds[0].(RecordEncounterJoinCmd)
	if join.WorldID != minasocoWorld || join.InstanceID != minasocoInst {
		t.Errorf("join world_id=%q instance_id=%q, want world %q instance %q",
			join.WorldID, join.InstanceID, minasocoWorld, minasocoInst)
	}
	leave := leaveCmds[0].(RecordEncounterLeaveCmd)
	if leave.VRCUserID != otherUser {
		t.Fatalf("leave user = %q, want %q", leave.VRCUserID, otherUser)
	}
	if leave.WorldID != minasocoWorld || leave.InstanceID != minasocoInst {
		t.Errorf("leave world_id=%q instance_id=%q, want world %q instance %q",
			leave.WorldID, leave.InstanceID, minasocoWorld, minasocoInst)
	}
}

func TestSessionCorrelator_Leave_afterDestinationBeforeJoin_usesLastSessionNotPending(t *testing.T) {
	base := time.Date(2026, 3, 22, 14, 20, 45, 0, time.UTC)
	const homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
	oldInst := homeWorld + ":85625~private(" + testvrc.PlayerUserID + ")~region(jp)"
	nextInst := homeWorld + ":62566~private(" + testvrc.PlayerUserID + ")~region(jp)"
	buddy := testvrc.PlayerUserID

	c := &SessionCorrelator{}
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: oldInst, OccurredAt: base})
	c.Apply(&EncounterEvent{
		VRCUserID:     buddy,
		DisplayName:   testvrc.PlayerDisplayName,
		Action:        EncounterActionJoin,
		EncounteredAt: base,
	})
	c.Apply(&DestinationSetEvent{
		WorldID:      homeWorld,
		FullInstance: nextInst,
		OccurredAt:   base,
	})
	c.Apply(&SessionEvent{Type: SessionEventEnd, OccurredAt: base})
	leaveCmds := c.Apply(&EncounterEvent{
		VRCUserID:     buddy,
		DisplayName:   testvrc.PlayerDisplayName,
		Action:        EncounterActionLeave,
		EncounteredAt: base.Add(time.Millisecond),
	})

	leave := leaveCmds[0].(RecordEncounterLeaveCmd)
	if leave.WorldID != homeWorld || leave.InstanceID != oldInst {
		t.Errorf("leave world_id=%q instance_id=%q, want world %q instance %q",
			leave.WorldID, leave.InstanceID, homeWorld, oldInst)
	}
}

func TestSessionCorrelator_SessionStart_emitsOrderedLifecycleCommands(t *testing.T) {
	base := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)
	c := &SessionCorrelator{}
	cmds := c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: testFullInstance, OccurredAt: base})
	if len(cmds) != 3 {
		t.Fatalf("commands = %d, want 3: %+v", len(cmds), cmds)
	}
	if _, ok := cmds[0].(EndPlaySessionCmd); !ok {
		t.Errorf("cmds[0] = %T, want EndPlaySessionCmd", cmds[0])
	}
	if _, ok := cmds[1].(CloseOpenEncountersAtCmd); !ok {
		t.Errorf("cmds[1] = %T, want CloseOpenEncountersAtCmd", cmds[1])
	}
	if start, ok := cmds[2].(StartPlaySessionCmd); !ok || start.InstanceID != testFullInstance {
		t.Errorf("cmds[2] = %+v, want StartPlaySessionCmd instance %q", cmds[2], testFullInstance)
	}
}

func TestSessionCorrelator_SessionStartEmptyInstanceIgnored(t *testing.T) {
	c := &SessionCorrelator{}
	if cmds := c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: "", OccurredAt: time.Now()}); cmds != nil {
		t.Fatalf("commands = %+v, want nil", cmds)
	}
}

func TestSessionCorrelator_UnhandledParsedEventsReturnNil(t *testing.T) {
	c := &SessionCorrelator{}
	base := time.Now()
	if cmds := c.Apply(&AvatarSwitchEvent{DisplayName: "A", AvatarName: "B", OccurredAt: base}); cmds != nil {
		t.Fatalf("AvatarSwitch commands = %+v, want nil", cmds)
	}
	if cmds := c.Apply(&VideoPlaybackEvent{URL: "https://example.com", OccurredAt: base}); cmds != nil {
		t.Fatalf("VideoPlayback commands = %+v, want nil", cmds)
	}
}

// Regression for GitHub bug report (2026-06-24): log-replayed home→cozy transition must not
// write Cozy with. onto the home world_id.
func TestSessionCorrelator_logReplay_homeToCozyTransition_roomNamesNotCrossAssigned(t *testing.T) {
	t.Setenv("TZ", "UTC")
	parser := NewLogParser()
	c := &SessionCorrelator{}

	const (
		homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
		cozyWorld = "wrld_6041ba53-0ac0-4b5b-9ecb-890ea2b0aefa"
	)
	homeInst := homeWorld + ":95147~private(" + testvrc.PlayerUserID + ")~region(jp)"
	cozyInst := cozyWorld + ":48580~friends(" + testvrc.FriendsHostUserID + ")~region(jp)"

	lines := []string{
		"2026.06.24 08:25:00 Debug      -  [Behaviour] Destination set: " + homeInst,
		"2026.06.24 08:25:01 Debug      -  [Behaviour] OnLeftRoom",
		"2026.06.24 08:25:03 Debug      -  [Behaviour] Entering Room: ホームチェックv6․0",
		"2026.06.24 08:25:03 Debug      -  [Behaviour] Joining " + homeInst,
		"2026.06.24 08:26:40 Debug      -  [Behaviour] Destination set: " + cozyInst,
		"2026.06.24 08:26:41 Debug      -  [Behaviour] OnLeftRoom",
		"2026.06.24 08:26:44 Debug      -  [Behaviour] Entering Room: Cozy with․",
		"2026.06.24 08:26:44 Debug      -  [Behaviour] Joining " + cozyInst,
	}

	var roomUpserts []UpsertWorldRoomNameCmd
	for _, line := range lines {
		base := ParseVRChatTimestamp(line, time.Time{})
		events, err := parser.ParseLine(line, base)
		if err != nil {
			t.Fatalf("ParseLine: %v", err)
		}
		for _, ev := range events {
			for _, cmd := range c.Apply(ev) {
				if room, ok := cmd.(UpsertWorldRoomNameCmd); ok {
					roomUpserts = append(roomUpserts, room)
				}
			}
		}
	}

	if len(roomUpserts) != 2 {
		t.Fatalf("room upserts = %d, want 2: %+v", len(roomUpserts), roomUpserts)
	}
	if roomUpserts[0].WorldID != homeWorld || roomUpserts[0].RoomName != "ホームチェックv6․0" {
		t.Fatalf("first room upsert = %+v, want home world name", roomUpserts[0])
	}
	if roomUpserts[1].WorldID != cozyWorld || roomUpserts[1].RoomName != "Cozy with․" {
		t.Fatalf("second room upsert = %+v, want cozy world name (must not target %s)", roomUpserts[1], homeWorld)
	}
}
