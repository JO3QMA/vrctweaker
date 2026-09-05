package media

import (
	"testing"
	"time"
)

func TestCorrelateActivity_singleSession(t *testing.T) {
	at := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	end := at.Add(time.Hour)
	sessions := []SessionAtTime{{
		InstanceID: "wrld_a:1~hidden(usr_x)~region(jp)",
		WorldID:    "wrld_a",
		StartTime:  at.Add(-time.Minute),
		EndTime:    &end,
	}}
	encounters := []EncounterAtTime{
		{VRCUserID: "usr_1", DisplayName: "Alice", InstanceID: sessions[0].InstanceID, JoinedAt: at.Add(-time.Minute)},
		{VRCUserID: "usr_2", DisplayName: "Bob", InstanceID: sessions[0].InstanceID, JoinedAt: at.Add(-30 * time.Second), LeftAt: timePtr(at.Add(30 * time.Minute))},
	}
	got := CorrelateActivity(at, "wrld_a", "wrld_a", "", sessions, encounters)
	if got.Status != EnrichmentStatusSuccess {
		t.Fatalf("status = %q", got.Status)
	}
	if got.InstanceID != sessions[0].InstanceID {
		t.Fatalf("instance = %q", got.InstanceID)
	}
	if len(got.Participants) != 2 {
		t.Fatalf("participants = %d", len(got.Participants))
	}
}

func TestCorrelateActivity_ambiguous(t *testing.T) {
	at := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	sessions := []SessionAtTime{
		{InstanceID: "wrld_a:1", WorldID: "wrld_a", StartTime: at.Add(-time.Hour)},
		{InstanceID: "wrld_a:2", WorldID: "wrld_a", StartTime: at.Add(-time.Hour)},
	}
	got := CorrelateActivity(at, "wrld_a", "wrld_a", "", sessions, nil)
	if got.Status != EnrichmentStatusAmbiguous {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestCorrelateActivity_conflictWorldID(t *testing.T) {
	at := time.Now()
	got := CorrelateActivity(at, "wrld_a", "wrld_b", "", nil, nil)
	if got.Status != EnrichmentStatusConflict {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestCorrelateActivity_conflictInstanceID(t *testing.T) {
	at := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	sessions := []SessionAtTime{{
		InstanceID: "wrld_a:2",
		WorldID:    "wrld_a",
		StartTime:  at.Add(-time.Minute),
	}}
	got := CorrelateActivity(at, "wrld_a", "wrld_a", "wrld_a:1", sessions, nil)
	if got.Status != EnrichmentStatusConflict {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestSnapshotParticipants_excludesLeft(t *testing.T) {
	at := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	left := at.Add(-time.Minute)
	encounters := []EncounterAtTime{
		{VRCUserID: "usr_gone", DisplayName: "Gone", InstanceID: "inst", JoinedAt: at.Add(-2 * time.Hour), LeftAt: &left},
		{VRCUserID: "usr_here", DisplayName: "Here", InstanceID: "inst", JoinedAt: at.Add(-time.Minute)},
	}
	got := SnapshotParticipants(encounters, "inst", at)
	if len(got) != 1 || got[0].VRCUserID != "usr_here" {
		t.Fatalf("got %+v", got)
	}
}

func TestIsEligibleForEnrichment(t *testing.T) {
	if !IsEligibleForEnrichment("wrld_a", true, EnrichmentFields{}, nil) {
		t.Fatal("VRC camera photo should be eligible")
	}
	if IsEligibleForEnrichment("wrld_a", true, EnrichmentFields{InstanceID: "x"}, &ScreenshotEnrichment{Status: EnrichmentStatusSuccess}) {
		t.Fatal("success with instance should not be eligible")
	}
	if !IsEligibleForEnrichment("", true, EnrichmentFields{}, nil) {
		t.Fatal("missing world_id should be eligible")
	}
}

func timePtr(t time.Time) *time.Time { return &t }
