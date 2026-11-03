package logwatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/usecase"
)

func TestProcessOutputLogFileFromOffset_stopsOnHandlerDispatchError(t *testing.T) {
	t.Setenv("TZ", "UTC")
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	line1 := "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining wrld_test:1~region(jp)\n"
	line2 := "2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_fail)\n"
	if err := writeTestFile(path, line1+line2); err != nil {
		t.Fatal(err)
	}

	encRepo := failSaveEncounterRepo{}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, encRepo, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	adapter := NewActivityIngestAdapter(uc, context.Background(), nil, nil, path)

	pos, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, 0, activity.NewLogParser(), adapter, nil, nil,
	)
	if err == nil {
		t.Fatal("expected dispatch error")
	}
	if !errors.Is(err, ErrHandlerDispatch) {
		t.Fatalf("err = %v, want ErrHandlerDispatch", err)
	}
	if pos != int64(len(line1)) {
		t.Fatalf("pos = %d, want %d (stop before failing line)", pos, len(line1))
	}
}

type failSaveEncounterRepo struct{ stubEncounterRepo }

func (failSaveEncounterRepo) Save(context.Context, *activity.UserEncounter) error {
	return errors.New("save fail")
}

func TestProcessOutputLogFileFromOffset_processedOffsetIgnoresGrowthAfterEOF(t *testing.T) {
	t.Setenv("TZ", "UTC")
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	line1 := "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining wrld_test:1~region(jp)\n"
	line2 := "2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_late)\n"
	if err := writeTestFile(path, line1); err != nil {
		t.Fatal(err)
	}

	pos, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, 0, activity.NewLogParser(), testEventHandler(func(activity.ParsedEvent) {}), nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := appendToTestFile(path, line2); err != nil {
		t.Fatal(err)
	}
	st, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if pos != int64(len(line1)) {
		t.Fatalf("pos = %d, want %d", pos, len(line1))
	}
	if st.Size() <= pos {
		t.Fatalf("file size = %d, want greater than processed pos %d", st.Size(), pos)
	}
}

func TestActivityIngestAdapter_notifySkippedWhenApplyCommandFails(t *testing.T) {
	ctx := context.Background()
	var notify int
	encRepo := &errEncounterRepo{}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, encRepo, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	a := NewActivityIngestAdapter(uc, ctx, nil, func() { notify++ }, testLogSource)
	a.Handle(&activity.SessionEvent{
		Type:       activity.SessionEventStart,
		InstanceID: testFullInstance,
		OccurredAt: time.Now(),
	})
	if err := a.Handle(&activity.EncounterEvent{
		VRCUserID: "usr_x", DisplayName: "X", Action: activity.EncounterActionJoin, EncounteredAt: time.Now(),
	}); err == nil {
		t.Fatal("expected ApplyCommand error")
	}
	if notify != 0 {
		t.Fatalf("notify calls = %d, want 0 on failed ingest", notify)
	}
}
