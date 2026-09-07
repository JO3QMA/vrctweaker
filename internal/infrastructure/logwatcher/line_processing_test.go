package logwatcher

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

func Test_parseOutputLogLine_nilParser(t *testing.T) {
	t.Parallel()
	_, _, err := parseOutputLogLine("line", nil, time.Now())
	if !errors.Is(err, errNilLineProcessorArg) {
		t.Fatalf("err = %v, want %v", err, errNilLineProcessorArg)
	}
}

func Test_parseOutputLogLine_parsesEncounter(t *testing.T) {
	t.Setenv("TZ", "UTC")
	line := "2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_abc)"
	events, baseTime, err := parseOutputLogLine(line, activity.NewLogParser(), time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Kind() != activity.EventKindEncounter {
		t.Fatalf("kind = %v", events[0].Kind())
	}
	want := time.Date(2026, 3, 21, 11, 32, 16, 0, time.Local)
	if !baseTime.Equal(want) {
		t.Fatalf("baseTime = %v, want %v", baseTime, want)
	}
}

func TestLineProcessor_nilHandler(t *testing.T) {
	t.Parallel()
	processor := NewLineProcessor(activity.NewLogParser(), nil)
	_, err := processor.Process("line")
	if !errors.Is(err, errNilLineProcessorArg) {
		t.Fatalf("err = %v, want %v", err, errNilLineProcessorArg)
	}
}

func TestLineProcessor_dispatchesEvents(t *testing.T) {
	t.Parallel()
	line := "2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_abc)"
	var got activity.ParsedEvent
	processor := NewLineProcessor(activity.NewLogParser(), FuncEventHandler(func(ev activity.ParsedEvent) {
		got = ev
	}))
	if _, err := processor.Process(line); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Kind() != activity.EventKindEncounter {
		t.Fatalf("unexpected event: %v", got)
	}
}

func TestScanOutputLogFromReader_respectsEndOffset(t *testing.T) {
	t.Parallel()
	content := "line1\nline2\nline3\n"
	br := bufio.NewReader(strings.NewReader(content))
	var lines []string
	pos, err := scanOutputLogFromReader(context.Background(), br, 0, int64(len("line1\n")), func(line ScannedLine) error {
		lines = append(lines, line.Trimmed)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if pos != int64(len("line1\n")) {
		t.Fatalf("pos = %d", pos)
	}
	if len(lines) != 1 || lines[0] != "line1" {
		t.Fatalf("lines = %v", lines)
	}
}

func TestScanOutputLogFromReader_invokesCallbackForEmptyLines(t *testing.T) {
	t.Parallel()
	content := "a\n\nb\n"
	br := bufio.NewReader(strings.NewReader(content))
	var trimmed []string
	_, err := scanOutputLogFromReader(context.Background(), br, 0, 0, func(line ScannedLine) error {
		trimmed = append(trimmed, line.Trimmed)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(trimmed) != 3 {
		t.Fatalf("callbacks = %d, want 3: %v", len(trimmed), trimmed)
	}
	if trimmed[0] != "a" || trimmed[1] != "" || trimmed[2] != "b" {
		t.Fatalf("trimmed = %v", trimmed)
	}
}

func TestFanoutHandler_dispatchesToAll(t *testing.T) {
	t.Parallel()
	var count int
	h := FanoutHandler{
		FuncEventHandler(func(activity.ParsedEvent) { count++ }),
		FuncEventHandler(func(activity.ParsedEvent) { count++ }),
	}
	h.Handle(&activity.EncounterEvent{Action: activity.EncounterActionJoin})
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}
