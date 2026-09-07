package logwatcher

import (
	"vrchat-tweaker/internal/domain/activity"
)

// EventHandler receives parsed events from the watcher.
type EventHandler interface {
	Handle(event activity.ParsedEvent)
}

// FuncEventHandler adapts a function to EventHandler.
type FuncEventHandler func(event activity.ParsedEvent)

// Handle implements EventHandler.
func (f FuncEventHandler) Handle(event activity.ParsedEvent) {
	if f != nil {
		f(event)
	}
}

// FanoutHandler is an exported slice of EventHandlers used to fan out parsed events
// (e.g. activity ingest + automation triggers) without a bespoke composite type.
type FanoutHandler []EventHandler

// Handle implements EventHandler.
// Each non-nil handler is invoked in order. Handlers must be safe for concurrent use
// from other goroutines; FanoutHandler does not synchronize between them.
func (h FanoutHandler) Handle(event activity.ParsedEvent) {
	for _, handler := range h {
		if handler != nil {
			handler.Handle(event)
		}
	}
}
