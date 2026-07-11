package logwatcher

import (
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/diag"
)

// Logger is a diagnostic logger for logwatcher (shared with picturewatcher via diag).
type Logger = diag.Logger

// EventHandler receives parsed events from the watcher.
type EventHandler interface {
	Handle(event activity.ParsedEvent)
}

// EventHandlerFunc adapts a function to EventHandler.
type EventHandlerFunc func(event activity.ParsedEvent)

// Handle implements EventHandler.
func (f EventHandlerFunc) Handle(event activity.ParsedEvent) {
	if f != nil {
		f(event)
	}
}

// LogParser parses a log line into events.
type LogParser interface {
	ParseLine(line string, baseTime time.Time) ([]activity.ParsedEvent, error)
}

const (
	pollInterval   = 500 * time.Millisecond
	readBufferSize = 64 * 1024
)
