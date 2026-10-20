package usecase

import (
	"encoding/json"
	"testing"
)

func TestPresenceChangeJSON_matchesFormerWailsDTO(t *testing.T) {
	section, err := json.Marshal(PresenceChangeSection{
		LoggedIn:          true,
		Status:            "active",
		StatusDescription: "hello",
		History:           []string{"a"},
	})
	if err != nil {
		t.Fatalf("marshal section: %v", err)
	}
	const wantSection = `{"loggedIn":true,"status":"active","statusDescription":"hello","history":["a"]}`
	if string(section) != wantSection {
		t.Fatalf("section json=%s", section)
	}

	applied, err := json.Marshal(PresenceChangeApplyResult{
		Status:            "busy",
		StatusDescription: "x",
	})
	if err != nil {
		t.Fatalf("marshal apply: %v", err)
	}
	const wantApply = `{"status":"busy","statusDescription":"x"}`
	if string(applied) != wantApply {
		t.Fatalf("apply json=%s", applied)
	}
}
