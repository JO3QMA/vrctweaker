package logwatcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/testvrc"
	"vrchat-tweaker/internal/usecase"
)

// Regression: checkpoint after Destination set must warm correlator so Entering Room
// in the resumed tail still upserts display name (pending from prefix).
func TestWarmSessionCorrelatorFromLogFile_restoresPendingDestinationBeforeResume(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")

	const cozyWorld = "wrld_6041ba53-0ac0-4b5b-9ecb-890ea2b0aefa"
	cozyInst := cozyWorld + ":48580~friends(" + testvrc.FriendsHostUserID + ")~region(jp)"

	prefix := []byte(
		"2026.06.24 08:26:40 Debug      -  [Behaviour] Destination set: " + cozyInst + "\n",
	)
	suffix := []byte(
		"2026.06.24 08:26:41 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:26:44 Debug      -  [Behaviour] Entering Room: Cozy with․\n",
	)
	full := append(append([]byte(nil), prefix...), suffix...)
	if writeErr := os.WriteFile(logPath, full, 0600); writeErr != nil {
		t.Fatal(writeErr)
	}

	worldRepo := &spyWorldInfoRepo{}
	uc := usecase.NewActivityUseCase(
		stubPlaySessionRepo{},
		stubEncounterRepo{},
		&fakeAppSettingsRepo{m: make(map[string]string)},
		nil,
		worldRepo,
	)
	parser := activity.NewLogParser()

	t.Run("without warm skips room name upsert", func(t *testing.T) {
		worldRepo.displayNameCalls = nil
		adapter := NewActivityIngestAdapter(uc, ctx, nil, nil)
		if _, err := ProcessOutputLogFileFromOffset(ctx, logPath, int64(len(prefix)), parser, adapter, nil, nil); err != nil {
			t.Fatal(err)
		}
		if len(worldRepo.displayNameCalls) != 0 {
			t.Fatalf("display name upserts = %d, want 0 without warm: %+v", len(worldRepo.displayNameCalls), worldRepo.displayNameCalls)
		}
	})

	t.Run("with warm upserts cozy display name", func(t *testing.T) {
		worldRepo.displayNameCalls = nil
		adapter := NewActivityIngestAdapter(uc, ctx, nil, nil)
		if err := WarmSessionCorrelatorFromLogFile(ctx, logPath, int64(len(prefix)), parser, adapter, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := ProcessOutputLogFileFromOffset(ctx, logPath, int64(len(prefix)), parser, adapter, nil, nil); err != nil {
			t.Fatal(err)
		}
		if len(worldRepo.displayNameCalls) != 1 {
			t.Fatalf("display name upserts = %d, want 1: %+v", len(worldRepo.displayNameCalls), worldRepo.displayNameCalls)
		}
		call := worldRepo.displayNameCalls[0]
		if call.worldID != cozyWorld || call.displayName != "Cozy with․" {
			t.Fatalf("upsert = %+v, want world %q name %q", call, cozyWorld, "Cozy with․")
		}
	})
}
