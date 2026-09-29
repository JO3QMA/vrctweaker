package usecase

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFindLocalAvatarCachePath_resolvesSymlinkRoot(t *testing.T) {
	root := t.TempDir()
	realCache := filepath.Join(root, "real")
	avatarID := "avtr_11111111-2222-3333-4444-555555555555"
	avatarDir := filepath.Join(realCache, avatarID)
	if err := os.MkdirAll(avatarDir, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(realCache, link); err != nil {
		t.Fatal(err)
	}

	got, err := FindLocalAvatarCachePath(link, avatarID)
	if err != nil {
		t.Fatal(err)
	}
	if got != avatarDir {
		t.Fatalf("got %q want %q", got, avatarDir)
	}
}

func TestFindLocalAvatarCachePath_findsDirectoryContainingAvatarID(t *testing.T) {
	root := t.TempDir()
	avatarID := "avtr_11111111-2222-3333-4444-555555555555"
	avatarDir := filepath.Join(root, "bundle", avatarID)
	if err := os.MkdirAll(filepath.Join(avatarDir, "data"), 0700); err != nil {
		t.Fatal(err)
	}

	got, err := FindLocalAvatarCachePath(root, avatarID)
	if err != nil {
		t.Fatal(err)
	}
	if got != avatarDir {
		t.Fatalf("got %q want %q", got, avatarDir)
	}
}

func TestFindLocalAvatarCachePath_ignoresPathContainingIDWithoutExactBase(t *testing.T) {
	root := t.TempDir()
	avatarID := "avtr_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	decoy := filepath.Join(root, "decoy_"+avatarID+"_suffix")
	if err := os.MkdirAll(decoy, 0700); err != nil {
		t.Fatal(err)
	}
	avatarDir := filepath.Join(root, avatarID)
	if err := os.MkdirAll(avatarDir, 0700); err != nil {
		t.Fatal(err)
	}

	got, err := FindLocalAvatarCachePath(root, avatarID)
	if err != nil {
		t.Fatal(err)
	}
	if got != avatarDir {
		t.Fatalf("got %q want %q", got, avatarDir)
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

func TestFindLocalAvatarCachePath_truncatedWhenWalkLimitExceeded(t *testing.T) {
	root := t.TempDir()
	avatarID := "avtr_11111111-2222-3333-4444-555555555555"
	for i := 0; i < 6; i++ {
		if err := os.MkdirAll(filepath.Join(root, fmt.Sprintf("noise%04d", i)), 0700); err != nil {
			t.Fatal(err)
		}
	}
	got, err := findLocalAvatarCachePathWithLimit(root, avatarID, 3)
	if !errors.Is(err, ErrAvatarCacheWalkTruncated) {
		t.Fatalf("err = %v want ErrAvatarCacheWalkTruncated", err)
	}
	if got != "" {
		t.Fatalf("got %q want empty", got)
	}

	target := filepath.Join(root, "zzz", avatarID)
	if mkdirErr := os.MkdirAll(target, 0700); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	got, err = findLocalAvatarCachePathWithLimit(root, avatarID, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Fatalf("got %q want %q", got, target)
	}
}
