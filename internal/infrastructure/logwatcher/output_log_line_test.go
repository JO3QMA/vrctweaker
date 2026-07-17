package logwatcher

import (
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
)

func TestDispatchOutputLogLine_nilParser(t *testing.T) {
	t.Parallel()
	called := false
	_, err := dispatchOutputLogLine("line", nil, testEventHandler(func(activity.ParsedEvent) {
		called = true
	}))
	if !errors.Is(err, errNilDispatchArg) {
		t.Fatalf("err = %v, want %v", err, errNilDispatchArg)
	}
	if called {
		t.Fatal("handler must not be called")
	}
}

func TestDispatchOutputLogLine_nilHandler(t *testing.T) {
	t.Parallel()
	_, err := dispatchOutputLogLine("line", activity.NewLogParser(), nil)
	if !errors.Is(err, errNilDispatchArg) {
		t.Fatalf("err = %v, want %v", err, errNilDispatchArg)
	}
}
