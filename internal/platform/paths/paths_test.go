package paths

import (
	"path/filepath"
	"testing"
)

func TestVRChatConfigPathOrFallback_usesDataDir(t *testing.T) {
	t.Parallel()
	dir, err := VRChatDataDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "config.json")
	got := VRChatConfigPathOrFallback()
	if got != want {
		t.Fatalf("VRChatConfigPathOrFallback() = %q, want %q", got, want)
	}
}

func TestDefaultVRChatOutputLogDir_matchesConfigParent(t *testing.T) {
	t.Parallel()
	dir, err := VRChatDataDir()
	if err != nil {
		t.Fatal(err)
	}
	if got := DefaultVRChatOutputLogDir(); got != dir {
		t.Fatalf("DefaultVRChatOutputLogDir() = %q, want %q", got, dir)
	}
}
