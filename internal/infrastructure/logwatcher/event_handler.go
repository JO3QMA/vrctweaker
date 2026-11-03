package logwatcher

import (
	"vrchat-tweaker/internal/domain/activity"
)

// EventHandler receives parsed events from the watcher.
type EventHandler interface {
	Handle(event activity.ParsedEvent) error
}

// FuncEventHandler adapts a function to EventHandler.
type FuncEventHandler func(event activity.ParsedEvent)

// Handle implements EventHandler.
func (f FuncEventHandler) Handle(event activity.ParsedEvent) error {
	if f != nil {
		f(event)
	}
	return nil
}

// FanoutHandler is an exported slice of EventHandlers used to fan out parsed events
// (e.g. activity ingest + automation triggers) without a bespoke composite type.
type FanoutHandler []EventHandler

// Handle implements EventHandler.
// Each non-nil handler is invoked in order. Handlers must be safe for concurrent use
// from other goroutines; FanoutHandler does not synchronize between them.
func (h FanoutHandler) Handle(event activity.ParsedEvent) error {
	var firstErr error
	for _, handler := range h {
		if handler != nil {
			if err := handler.Handle(event); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
