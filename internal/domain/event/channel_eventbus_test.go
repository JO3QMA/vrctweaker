package event

import (
	"context"
	"errors"
	"testing"
)

func TestChannelEventBus_PublishSubscribe(t *testing.T) {
	bus := NewChannelEventBus()
	ctx := context.Background()
	var got []*Event
	unsub := bus.Subscribe("topic", func(_ context.Context, e *Event) error {
		got = append(got, e)
		return nil
	})
	defer unsub()

	ev := &Event{Type: "ping", Payload: "hello"}
	if err := bus.Publish(ctx, "topic", ev); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != ev {
		t.Fatalf("got events = %+v", got)
	}
}

func TestChannelEventBus_UnsubscribeStopsDelivery(t *testing.T) {
	bus := NewChannelEventBus()
	ctx := context.Background()
	calls := 0
	unsub := bus.Subscribe("topic", func(context.Context, *Event) error {
		calls++
		return nil
	})
	unsub()
	if err := bus.Publish(ctx, "topic", &Event{Type: "x"}); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("calls after unsubscribe = %d", calls)
	}
}

func TestChannelEventBus_HandlerErrorDoesNotStopOthers(t *testing.T) {
	bus := NewChannelEventBus()
	ctx := context.Background()
	second := false
	bus.Subscribe("topic", func(context.Context, *Event) error {
		return errors.New("fail")
	})
	bus.Subscribe("topic", func(context.Context, *Event) error {
		second = true
		return nil
	})
	_ = bus.Publish(ctx, "topic", &Event{Type: "x"})
	if !second {
		t.Fatal("second handler should still run")
	}
}
