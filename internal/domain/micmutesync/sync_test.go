package micmutesync

import (
	"testing"
	"time"
)

func TestEchoGuard_SuppressesOppositeSource(t *testing.T) {
	var g EchoGuard
	g.Suppress(SourceVRChat, time.Second)
	if g.ShouldIgnore(SourceVRChat) {
		t.Fatal("same source should not be ignored")
	}
	if !g.ShouldIgnore(SourceDiscord) {
		t.Fatal("opposite source should be ignored during suppression")
	}
	time.Sleep(1100 * time.Millisecond)
	if g.ShouldIgnore(SourceDiscord) {
		t.Fatal("suppression should expire")
	}
}

func TestNeedsMuteToggle(t *testing.T) {
	if NeedsMuteToggle(false, false, true) {
		// unknown current -> toggle once
	} else {
		t.Fatal("unknown current should toggle")
	}
	if NeedsMuteToggle(true, true, true) {
		t.Fatal("already muted")
	}
	if !NeedsMuteToggle(true, false, true) {
		t.Fatal("should toggle to mute")
	}
}
