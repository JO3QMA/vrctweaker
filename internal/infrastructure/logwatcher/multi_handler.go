package logwatcher

import (
	"vrchat-tweaker/internal/domain/activity"
)

// MultiHandler runs multiple EventHandlers for each event.
type MultiHandler struct {
	handlers []EventHandler
}

// NewMultiHandler creates a handler that delegates to all given handlers.
func NewMultiHandler(handlers ...EventHandler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Handle implements EventHandler.
func (m *MultiHandler) Handle(event activity.ParsedEvent) {
	for _, h := range m.handlers {
		h.Handle(event)
	}
}
