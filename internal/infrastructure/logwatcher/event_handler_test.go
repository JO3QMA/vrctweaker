package logwatcher

import (
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
)

func TestFanoutHandler_stopsOnFirstError(t *testing.T) {
	t.Parallel()
	var secondCalls int
	errHandler := failEventHandler{err: errors.New("ingest fail")}
	okHandler := FuncEventHandler(func(activity.ParsedEvent) {
		secondCalls++
	})
	h := FanoutHandler{errHandler, okHandler}
	ev := &activity.EncounterEvent{Action: activity.EncounterActionJoin}
	if err := h.Handle(ev); err == nil {
		t.Fatal("expected error")
	}
	if secondCalls != 0 {
		t.Fatalf("second handler calls = %d, want 0", secondCalls)
	}
}

type failEventHandler struct{ err error }

func (f failEventHandler) Handle(activity.ParsedEvent) error { return f.err }
