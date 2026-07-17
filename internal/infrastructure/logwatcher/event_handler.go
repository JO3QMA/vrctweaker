package logwatcher

import (
	"vrchat-tweaker/internal/domain/activity"
)

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
