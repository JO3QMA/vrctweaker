package micmutesync

import (
	"strings"
	"testing"
)

func TestEnsureOSCInLaunchArgs_addsDefault(t *testing.T) {
	got := EnsureOSCInLaunchArgs("-no-vr", "")
	if !strings.Contains(got, "--osc=9000:127.0.0.1:9001") {
		t.Fatalf("got %q", got)
	}
}

func TestEnsureOSCInLaunchArgs_preservesExisting(t *testing.T) {
	in := "--osc=9100:127.0.0.1:9101"
	if EnsureOSCInLaunchArgs(in, "") != in {
		t.Fatal("should not change existing osc")
	}
}
