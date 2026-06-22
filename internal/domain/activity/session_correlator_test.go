package activity

import (
	"testing"
	"time"
)

const testWorldID = "wrld_beddab1e-fee1-cafe-f00d-ca7c0dd1eca7"

var testFullInstance = testWorldID + ":41550~hidden(usr_aeab2f4d-40b4-4f73-acbd-608ac47b763e)~region(jp)"

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
	prevInst := prevWorld + ":98225~hidden(usr_83ba5dc2-2912-4a21-a514-8b954e60a79b)~region(jp)"
	const nextWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
	nextInst := nextWorld + ":77788~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"

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
	homeInst := homeWorld + ":04910~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"
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

func TestSessionCorrelator_OtherPlayerLeave_afterJoining_keepsWorldContext(t *testing.T) {
	base := time.Date(2026, 3, 18, 0, 1, 0, 0, time.UTC)
	const minasocoWorld = "wrld_c03f8195-3c64-46d8-b5ae-242f214c9404"
	minasocoInst := minasocoWorld + ":98225~hidden(usr_83ba5dc2-2912-4a21-a514-8b954e60a79b)~region(jp)"
	const otherUser = "usr_1564b5c1-888a-4d08-b7f4-dcedcf702a90"

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
	oldInst := homeWorld + ":85625~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"
	nextInst := homeWorld + ":62566~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"
	const buddy = "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e"

	c := &SessionCorrelator{}
	c.Apply(&SessionEvent{Type: SessionEventStart, InstanceID: oldInst, OccurredAt: base})
	c.Apply(&EncounterEvent{
		VRCUserID:     buddy,
		DisplayName:   "ぶっちゃん！",
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
		DisplayName:   "ぶっちゃん！",
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
