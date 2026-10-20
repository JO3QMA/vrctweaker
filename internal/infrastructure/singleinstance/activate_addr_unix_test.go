//go:build !windows

package singleinstance

import (
	"fmt"
	"os"
	"testing"
)

func TestAbstractActivateAddrIncludesUIDAndFingerprint(t *testing.T) {
	fp := dataDirFingerprint("/tmp/example-app-data")
	addr := abstractActivateAddr("VRChatTweaker", fp)
	want := fmt.Sprintf("@VRChatTweaker_%d_%s_activate", os.Getuid(), fp)
	if addr != want {
		t.Fatalf("got %q want %q", addr, want)
	}
}
