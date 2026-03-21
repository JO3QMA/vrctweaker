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
			line:       "Joining wrld_abc123:12345",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventStart,
			wantInstID: "wrld_abc123:12345",
		},
		{
			name:      "Entering Room does not start session",
			line:      "2026.03.17 23:59:58 Debug      -  [Behaviour] Entering Room: My World",
			wantCount: 0,
		},
		{
			name:       "Joining wrld with full log prefix",
			line:       "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)",
			wantCount:  1,
			wantKind:   EventKindSession,
			wantType:   SessionEventStart,
			wantInstID: "wrld_db637cfb-64f8-4109-977b-6b755482f133:88577",
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
