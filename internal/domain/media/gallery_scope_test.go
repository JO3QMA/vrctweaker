package media

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPictureFolderPathPrefix(t *testing.T) {
	root := filepath.Join("C:", "Pictures", "VRChat")
	got := PictureFolderPathPrefix(root)
	want := root + string(filepath.Separator)
	if got != want {
		t.Fatalf("PictureFolderPathPrefix() = %q, want %q", got, want)
	}

	if PictureFolderPathPrefix("") != "" {
		t.Fatal("empty root should yield empty prefix")
	}
}

func TestPictureFolderPathPrefix_doesNotMatchSiblingFolderNames(t *testing.T) {
	root := filepath.Join("C:", "Pictures", "VRChat")
	prefix := filepath.ToSlash(PictureFolderPathPrefix(root))
	sibling := filepath.ToSlash(filepath.Join("C:", "Pictures", "VRChat_old", "shot.png"))
	if strings.HasPrefix(sibling, prefix) {
		t.Fatalf("prefix %q should not match sibling path %q", prefix, sibling)
	}
	inScope := filepath.ToSlash(filepath.Join(root, "shot.png"))
	if !strings.HasPrefix(inScope, prefix) {
		t.Fatalf("prefix %q should match in-scope path %q", prefix, inScope)
	}
}
