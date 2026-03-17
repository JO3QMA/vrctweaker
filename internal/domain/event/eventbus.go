package event

import "context"

// Event represents a domain event payload.
type Event struct {
	Type    string
	Payload interface{}
}

// EventBus provides pub/sub messaging between domains.
type EventBus interface {
	// Publish sends an event to all subscribers of the given topic.
	Publish(ctx context.Context, topic string, event *Event) error
	// Subscribe registers a handler for the given topic. Returns a function to unsubscribe.
	Subscribe(topic string, handler func(ctx context.Context, event *Event) error) func()
}
