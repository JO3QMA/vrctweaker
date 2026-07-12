//go:build windows

package usecase

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNeedsOfficialLink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache", "yt-dlp.exe")
	tools := filepath.Join(dir, "Tools", "yt-dlp.exe")
	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(tools), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte("official"), 0o755); err != nil {
		t.Fatal(err)
	}

	need, err := NeedsOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Fatal("missing tools should need link")
	}

	if writeErr := os.WriteFile(tools, []byte("bundled"), 0o755); writeErr != nil {
		t.Fatal(writeErr)
	}
	need, err = NeedsOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Fatal("plain file should need link")
	}

	if rmErr := os.Remove(tools); rmErr != nil {
		t.Fatal(rmErr)
	}
	if linkErr := os.Symlink(cache, tools); linkErr != nil {
		t.Fatal(linkErr)
	}
	need, err = NeedsOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatal("correct symlink should not need link")
	}

	other := filepath.Join(dir, "other.exe")
	if writeErr := os.WriteFile(other, []byte("x"), 0o755); writeErr != nil {
		t.Fatal(writeErr)
	}
	if rmErr := os.Remove(tools); rmErr != nil {
		t.Fatal(rmErr)
	}
	if linkErr := os.Symlink(other, tools); linkErr != nil {
		t.Fatal(linkErr)
	}
	need, err = NeedsOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Fatal("wrong symlink target should need link")
	}
}

func TestNeedsOfficialLink_missingCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	need, err := NeedsOfficialLink(filepath.Join(dir, "yt-dlp.exe"), filepath.Join(dir, "missing.exe"))
	if !errors.Is(err, ErrYTDLPCacheMissing) {
		t.Fatalf("got need=%v err=%v", need, err)
	}
	if need {
		t.Fatal("missing cache should not request link")
	}
}

func TestLinkToolsToCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache", "yt-dlp.exe")
	tools := filepath.Join(dir, "Tools", "yt-dlp.exe")
	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte("official"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(tools), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tools, []byte("bundled"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := LinkToolsToCache(tools, cache, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	need, err := NeedsOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatal("after link, should be effective")
	}
}

func TestOfficialYTDLPCachePath(t *testing.T) {
	t.Parallel()
	local := filepath.Join(string(filepath.Separator), "Users", "x", "AppData", "Local")
	p := officialYTDLPCachePathFromLocal(local)
	if filepath.Base(p) != "yt-dlp.exe" {
		t.Fatalf("base: %s", p)
	}
	wantSuffix := filepath.ToSlash(filepath.Join("vrchat-tweaker", "ytdlp"))
	if !strings.HasSuffix(filepath.ToSlash(filepath.Dir(p)), wantSuffix) {
		t.Fatalf("dir: %s", p)
	}
	tools := vrchatYTDLPToolsPathFromLocal(local)
	if !strings.Contains(filepath.ToSlash(tools), "/LocalLow/VRChat/VRChat/Tools/yt-dlp.exe") &&
		!strings.Contains(filepath.ToSlash(tools), "LocalLow/VRChat/VRChat/Tools/yt-dlp.exe") {
		t.Fatalf("tools: %s", tools)
	}
}
