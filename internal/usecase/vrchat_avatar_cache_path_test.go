package usecase

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLocalAvatarCachePath_findsDirectoryContainingAvatarID(t *testing.T) {
	root := t.TempDir()
	avatarID := "avtr_11111111-2222-3333-4444-555555555555"
	target := filepath.Join(root, "bundle", avatarID, "data")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}

	got, err := FindLocalAvatarCachePath(root, avatarID)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Fatalf("got %q want %q", got, target)
	}
}

func TestFindLocalAvatarCachePath_emptyWhenInvalidID(t *testing.T) {
	got, err := FindLocalAvatarCachePath(t.TempDir(), "not-an-id")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestFindLocalAvatarCachePath_emptyWhenMissing(t *testing.T) {
	got, err := FindLocalAvatarCachePath(t.TempDir(), "avtr_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q", got)
	}
}
