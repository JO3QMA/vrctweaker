package event

import (
	"context"
	"sync"
)

type subscriber struct {
	handler func(context.Context, *Event) error
}

// ChannelEventBus is a channel-based implementation of EventBus.
type ChannelEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscriber
}

// NewChannelEventBus creates a new ChannelEventBus.
func NewChannelEventBus() *ChannelEventBus {
	return &ChannelEventBus{
		subscribers: make(map[string][]*subscriber),
	}
}

// Publish sends an event to all subscribers.
func (b *ChannelEventBus) Publish(ctx context.Context, topic string, event *Event) error {
	b.mu.RLock()
	subs := b.subscribers[topic]
	b.mu.RUnlock()

	for _, s := range subs {
		_ = s.handler(ctx, event)
	}
	return nil
}

// Subscribe registers a handler for a topic.
func (b *ChannelEventBus) Subscribe(topic string, handler func(context.Context, *Event) error) func() {
	s := &subscriber{handler: handler}
	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], s)
	idx := len(b.subscribers[topic]) - 1
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		curr := b.subscribers[topic]
		if idx < len(curr) && curr[idx] == s {
			b.subscribers[topic] = append(curr[:idx], curr[idx+1:]...)
		}
	}
}
