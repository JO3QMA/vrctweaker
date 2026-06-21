package discordrpc

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFrameWaiter_deliverAndWait(t *testing.T) {
	w := newFrameWaiter()
	nonce := "test-nonce"
	w.register(nonce)
	defer w.unregister(nonce)

	go func() {
		time.Sleep(10 * time.Millisecond)
		if !w.deliver(frameMessage{Nonce: nonce, Cmd: "PING"}) {
			t.Error("deliver failed")
		}
	}()

	msg, err := w.wait(nonce, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Cmd != "PING" {
		t.Fatalf("cmd %q", msg.Cmd)
	}
}

func TestFrameWaiter_deliverIgnoresUnknownNonce(t *testing.T) {
	w := newFrameWaiter()
	if w.deliver(frameMessage{Nonce: "missing", Cmd: "PING"}) {
		t.Fatal("expected false for unknown nonce")
	}
}

func TestFrameWaiter_waitTimeout(t *testing.T) {
	w := newFrameWaiter()
	w.register("late")
	defer w.unregister("late")
	_, err := w.wait("late", 20*time.Millisecond)
	if err == nil || err.Error() != "discord_rpc_timeout" {
		t.Fatalf("err=%v", err)
	}
}

func TestFrameWaiter_deliverVoiceEventWithoutNonce(t *testing.T) {
	w := newFrameWaiter()
	if w.deliver(frameMessage{
		Evt:  "VOICE_SETTINGS_UPDATE",
		Data: json.RawMessage(`{"mute":true}`),
	}) {
		t.Fatal("events without nonce should not be delivered to waiters")
	}
}
