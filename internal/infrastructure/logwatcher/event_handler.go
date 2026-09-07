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

// FanoutHandler dispatches each event to every non-nil handler in order.
type FanoutHandler []EventHandler

// Handle implements EventHandler.
func (h FanoutHandler) Handle(event activity.ParsedEvent) {
	for _, handler := range h {
		if handler != nil {
			handler.Handle(event)
		}
	}
}
