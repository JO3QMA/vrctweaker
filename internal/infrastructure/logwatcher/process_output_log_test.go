package logwatcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
)

func TestProcessOutputLogFile_DispatchesEvents(t *testing.T) {
	t.Setenv("TZ", "UTC")
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	content := "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)\n" +
		"2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_abc)\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	parser := activity.NewLogParser()
	var mu sync.Mutex
	var kinds []activity.EventKind
	h := EventHandlerFunc(func(ev activity.ParsedEvent) {
		mu.Lock()
		kinds = append(kinds, ev.Kind())
		mu.Unlock()
	})

	if err := ProcessOutputLogFile(context.Background(), path, parser, h, nil); err != nil {
		t.Fatalf("ProcessOutputLogFile: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(kinds) != 2 {
		t.Fatalf("got %d events, want 2: %v", len(kinds), kinds)
	}
	if kinds[0] != activity.EventKindSession || kinds[1] != activity.EventKindEncounter {
		t.Fatalf("unexpected kinds: %v", kinds)
	}
}
