package activity

import (
	"testing"
	"time"
)

func TestLogParser_ParseLine_Encounter(t *testing.T) {
	base := time.Date(2025, 3, 17, 12, 0, 0, 0, time.UTC)
	p := NewLogParser()

	tests := []struct {
		name     string
		line     string
		wantKind EventKind
		wantAct  string
		wantName string
		wantID   string
	}{
		{
			name:     "OnPlayerJoined with user ID",
			line:     "OnPlayerJoined Alice (usr_abc123)",
			wantKind: EventKindEncounter,
			wantAct:  EncounterActionJoin,
			wantName: "Alice",
			wantID:   "usr_abc123",
		},
		{
			name:     "OnPlayerJoined with prefix",
			line:     "[Time: 42.5] OnPlayerJoined Bob (usr_def456)",
			wantKind: EventKindEncounter,
			wantAct:  EncounterActionJoin,
			wantName: "Bob",
			wantID:   "usr_def456",
		},
		{
			name:     "OnPlayerJoined VRChat Behaviour line",
			line:     "2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined ぶっちゃん！ (usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)",
			wantKind: EventKindEncounter,
			wantAct:  EncounterActionJoin,
			wantName: "ぶっちゃん！",
			wantID:   "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e",
		},
		{
			name:     "OnPlayerLeft with user ID",
			line:     "OnPlayerLeft Charlie (usr_ghi789)",
			wantKind: EventKindEncounter,
			wantAct:  EncounterActionLeave,
			wantName: "Charlie",
			wantID:   "usr_ghi789",
		},
		{
			name:     "OnPlayerJoined display name only",
			line:     "OnPlayerJoined Some User Name",
			wantKind: EventKindEncounter,
			wantAct:  EncounterActionJoin,
			wantName: "Some User Name",
			wantID:   "",
		},
		{
			name:     "OnPlayerLeft display name only",
			line:     "OnPlayerLeft Another User",
			wantKind: EventKindEncounter,
			wantAct:  EncounterActionLeave,
			wantName: "Another User",
			wantID:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := p.ParseLine(tt.line, base)
			if err != nil {
				t.Fatalf("ParseLine() err = %v", err)
			}
			if len(events) != 1 {
				t.Fatalf("ParseLine() returned %d events, want 1", len(events))
			}
			e, ok := events[0].(*EncounterEvent)
			if !ok {
				t.Fatalf("ParseLine() event type = %T, want *EncounterEvent", events[0])
			}
			if e.Kind() != tt.wantKind {
				t.Errorf("Kind() = %v, want %v", e.Kind(), tt.wantKind)
			}
			if e.Action != tt.wantAct {
				t.Errorf("Action = %q, want %q", e.Action, tt.wantAct)
			}
			if e.DisplayName != tt.wantName {
				t.Errorf("DisplayName = %q, want %q", e.DisplayName, tt.wantName)
			}
			if e.VRCUserID != tt.wantID {
				t.Errorf("VRCUserID = %q, want %q", e.VRCUserID, tt.wantID)
			}
			if !e.EncounteredAt.Equal(base) {
				t.Errorf("EncounteredAt = %v, want %v", e.EncounteredAt, base)
			}
		})
	}
}

func TestParseVRChatTimestamp(t *testing.T) {
	t.Setenv("TZ", "UTC")
	fallback := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	line := "2026.03.17 23:59:58 Debug      -  [Behaviour] OnPlayerJoined x (usr_abc)"
	got := ParseVRChatTimestamp(line, fallback)
	want := time.Date(2026, 3, 17, 23, 59, 58, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("ParseVRChatTimestamp() = %v, want %v", got, want)
	}
	if !ParseVRChatTimestamp("no timestamp here", fallback).Equal(fallback) {
		t.Errorf("ParseVRChatTimestamp() should return fallback for bare line")
	}
}

func TestLogParser_ParseLine_Session(t *testing.T) {
	base := time.Date(2025, 3, 17, 12, 0, 0, 0, time.UTC)
	p := NewLogParser()

	tests := []struct {
		name       string
		line       string
		wantCount  int
		wantKind   EventKind
		wantType   string
		wantInstID string
	}{
		{
			name:       "Joining wrld",
			line:       "Joining wrld_abc123:12345~public",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventStart,
			wantInstID: "wrld_abc123:12345~public",
		},
		{
			name:       "Joining wrld with full log prefix",
			line:       "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventStart,
			wantInstID: "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)",
		},
		{
			name:       "OnPlayerLeftRoom",
			line:       "2026.03.18 00:04:09 Debug      -  [Behaviour] OnPlayerLeftRoom",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventEnd,
			wantInstID: "",
		},
		{
			name:       "OnLeftRoom",
			line:       "OnLeftRoom",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventEnd,
			wantInstID: "",
		},
		{
			name:       "Leaving room",
			line:       "Leaving room",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventEnd,
			wantInstID: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := p.ParseLine(tt.line, base)
			if err != nil {
				t.Fatalf("ParseLine() err = %v", err)
			}
			if len(events) != tt.wantCount {
				t.Fatalf("ParseLine() returned %d events, want %d", len(events), tt.wantCount)
			}
			if tt.wantCount == 0 {
				return
			}
			e, ok := events[0].(*SessionEvent)
			if !ok {
				t.Fatalf("ParseLine() event type = %T, want *SessionEvent", events[0])
			}
			if e.Kind() != tt.wantKind {
				t.Errorf("Kind() = %v, want %v", e.Kind(), tt.wantKind)
			}
			if e.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", e.Type, tt.wantType)
			}
			if e.InstanceID != tt.wantInstID {
				t.Errorf("InstanceID = %q, want %q", e.InstanceID, tt.wantInstID)
			}
		})
	}
}

func TestLogParser_ParseLine_Unparseable(t *testing.T) {
	p := NewLogParser()
	base := time.Now()

	unparseable := []string{
		"",
		"some random log line",
		"Loading level",
		"[Time: 1.0] Unrelated message",
	}
	for _, line := range unparseable {
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Errorf("ParseLine(%q) err = %v", line, err)
		}
		if len(events) != 0 {
			t.Errorf("ParseLine(%q) returned %d events, want 0", line, len(events))
		}
	}
}

func TestLogParser_ParseLine_DestinationRoomAvatarVideo(t *testing.T) {
	base := time.Date(2026, 3, 17, 23, 59, 56, 0, time.UTC)
	p := NewLogParser()

	t.Run("Destination set", func(t *testing.T) {
		line := "2026.03.17 23:59:56 Debug      -  [Behaviour] Destination set: wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b:64190~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("got %d events", len(events))
		}
		d, ok := events[0].(*DestinationSetEvent)
		if !ok {
			t.Fatalf("type %T", events[0])
		}
		if d.WorldID != "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b" || d.InstanceID != "64190" || d.InstanceType != "private" {
			t.Errorf("destination fields %+v", d)
		}
		if d.OwnerUserID != "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e" || d.Region != "jp" {
			t.Errorf("owner/region %+v", d)
		}
	})

	t.Run("Destination set group instance with groupAccessType before region", func(t *testing.T) {
		line := "2026.03.22 00:49:03 Debug      -  [Behaviour] Destination set: wrld_b2d24c29-1ded-4990-a90d-dd6dcc440300:mosco1~group(grp_55a159da-da85-4bf3-893d-65fc50abe6c1)~groupAccessType(public)~region(use)"
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("got %d events", len(events))
		}
		d, ok := events[0].(*DestinationSetEvent)
		if !ok {
			t.Fatalf("type %T", events[0])
		}
		wantWid := "wrld_b2d24c29-1ded-4990-a90d-dd6dcc440300"
		if d.WorldID != wantWid || d.InstanceID != "mosco1" {
			t.Errorf("destination fields %+v", d)
		}
		if d.Region != "use" {
			t.Errorf("Region = %q, want use", d.Region)
		}
		wantFull := wantWid + ":mosco1~group(grp_55a159da-da85-4bf3-893d-65fc50abe6c1)~groupAccessType(public)~region(use)"
		if d.FullInstance != wantFull {
			t.Errorf("FullInstance = %q", d.FullInstance)
		}
	})

	t.Run("Entering Room", func(t *testing.T) {
		line := "2026.03.17 23:59:58 Debug      -  [Behaviour] Entering Room: ホームチェックv6․0"
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("got %d events", len(events))
		}
		r, ok := events[0].(*RoomNameEvent)
		if !ok {
			t.Fatalf("type %T", events[0])
		}
		if r.RoomName != "ホームチェックv6․0" {
			t.Errorf("RoomName = %q", r.RoomName)
		}
	})

	t.Run("Avatar switch", func(t *testing.T) {
		line := "2026.03.18 00:00:08 Debug      -  [Behaviour] Switching ぶっちゃん！ to avatar RearAlice （SailorMaid）"
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("got %d events", len(events))
		}
		a, ok := events[0].(*AvatarSwitchEvent)
		if !ok {
			t.Fatalf("type %T", events[0])
		}
		if a.DisplayName != "ぶっちゃん！" || a.AvatarName != "RearAlice （SailorMaid）" {
			t.Errorf("avatar %+v", a)
		}
	})

	t.Run("Video playback", func(t *testing.T) {
		line := "2026.03.18 00:01:12 Debug      -  [Video Playback] Attempting to resolve URL 'https://youtu.be/-I1aPyp-_uE?si=x'"
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("got %d events", len(events))
		}
		v, ok := events[0].(*VideoPlaybackEvent)
		if !ok {
			t.Fatalf("type %T", events[0])
		}
		if v.URL != "https://youtu.be/-I1aPyp-_uE?si=x" {
			t.Errorf("URL = %q", v.URL)
		}
	})

	t.Run("Switching to network region ignored", func(t *testing.T) {
		line := "2026.03.17 23:59:57 Debug      -  [Behaviour] Switching to network region jp (current state: ConnectedToNameServer)"
		events, err := p.ParseLine(line, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 0 {
			t.Fatalf("want 0 events, got %v", events)
		}
	})
}
