package logwatcher

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/testvrc"
	"vrchat-tweaker/internal/usecase"
)

func TestTailIngestRetryDelay_capsAtMax(t *testing.T) {
	if d := tailIngestRetryDelay(0); d != tailIngestRetryInitial {
		t.Fatalf("attempt 0 = %v, want %v", d, tailIngestRetryInitial)
	}
	if d := tailIngestRetryDelay(20); d != tailIngestRetryMax {
		t.Fatalf("attempt 20 = %v, want max %v", d, tailIngestRetryMax)
	}
}

func TestTailOutputLogFile_retriesTransientIngestOnce(t *testing.T) {
	t.Setenv("TZ", "UTC")
	oldWait := tailIngestRetryWait
	tailIngestRetryWaitMu.Lock()
	tailIngestRetryWait = func(ctx context.Context, attempt int) bool {
		select {
		case <-ctx.Done():
			return true
		case <-time.After(2 * time.Millisecond):
			return false
		}
	}
	tailIngestRetryWaitMu.Unlock()
	t.Cleanup(func() {
		tailIngestRetryWaitMu.Lock()
		tailIngestRetryWait = oldWait
		tailIngestRetryWaitMu.Unlock()
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	const world = "wrld_beddab1e-fee1-cafe-f00d-ca7c0dd1eca7"
	inst := world + ":41550~hidden(" + testvrc.EmbedUserID + ")~region(jp)"
	content := "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining " + inst + "\n" +
		"2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_tail_retry)\n"
	if err := writeTestFile(path, content); err != nil {
		t.Fatal(err)
	}

	spy := &spyEncounterRepo{}
	repo := &failOnceSaveEncounterRepo{spyEncounterRepo: spy}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, repo, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	adapter := NewActivityIngestAdapter(uc, context.Background(), nil, nil, path)

	var (
		mu             sync.Mutex
		checkpoints    []int64
		lastCheckpoint int64
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		tailOutputLogFile(ctx, path, 0, activity.NewLogParser(), adapter, nil, tailCallbacks{
			checkpoint: func(offset int64, _ time.Time) {
				mu.Lock()
				checkpoints = append(checkpoints, offset)
				lastCheckpoint = offset
				if offset == int64(len(content)) {
					cancel()
				}
				mu.Unlock()
			},
		})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("tail did not finish after successful ingest")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(spy.saves) != 1 {
		t.Fatalf("Save calls = %d, want 1 (exactly once after retry)", len(spy.saves))
	}
	if lastCheckpoint != int64(len(content)) {
		t.Fatalf("checkpoint offset = %d, want %d", lastCheckpoint, len(content))
	}
	if len(checkpoints) < 2 {
		t.Fatalf("checkpoints = %d, want at least 2 (join line + encounter line)", len(checkpoints))
	}
}

type failOnceSaveEncounterRepo struct {
	*spyEncounterRepo
	saveFailed bool
}

func (r *failOnceSaveEncounterRepo) Save(ctx context.Context, e *activity.UserEncounter) error {
	if !r.saveFailed {
		r.saveFailed = true
		return errors.New("transient sqlite")
	}
	return r.spyEncounterRepo.Save(ctx, e)
}
