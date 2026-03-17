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

func TestLogParser_ParseLine_Session(t *testing.T) {
	base := time.Date(2025, 3, 17, 12, 0, 0, 0, time.UTC)
	p := NewLogParser()

	tests := []struct {
		name       string
		line       string
		wantKind   EventKind
		wantType   string
		wantInstID string
	}{
		{
			name:       "Joining wrld",
			line:       "Joining wrld_abc123:12345",
			wantKind:   EventKindSession,
			wantType:   SessionEventStart,
			wantInstID: "wrld_abc123:12345",
		},
		{
			name:       "Joining room",
			line:       "Joining or Creating Room",
			wantKind:   EventKindSession,
			wantType:   SessionEventStart,
			wantInstID: "",
		},
		{
			name:       "OnLeftRoom",
			line:       "OnLeftRoom",
			wantKind:   EventKindSession,
			wantType:   SessionEventEnd,
			wantInstID: "",
		},
		{
			name:       "Leaving room",
			line:       "Leaving room",
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
			if len(events) != 1 {
				t.Fatalf("ParseLine() returned %d events, want 1", len(events))
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
