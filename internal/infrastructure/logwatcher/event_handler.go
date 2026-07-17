package logwatcher

import (
	"vrchat-tweaker/internal/domain/activity"
)

// EventHandler receives parsed events from the watcher.
type EventHandler interface {
	Handle(event activity.ParsedEvent)
}
