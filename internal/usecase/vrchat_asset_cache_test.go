package usecase

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

type stubVRChatRunning struct {
	running bool
	err     error
}

func (s stubVRChatRunning) VRChatRunning() (bool, error) {
	return s.running, s.err
}

func TestClearVRChatAssetCache_clearsContents(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.bin"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "b.bin"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	uc := newTestAssetCacheUC(t, dir, "", stubVRChatRunning{})
	n, err := uc.Clear()
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if n != 2 {
		t.Fatalf("n = %d, want 2 top-level entries", n)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("dir still has %d entries", len(entries))
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("cache dir itself should remain: %v", err)
	}
}

func TestClearVRChatAssetCache_emptyDir(t *testing.T) {
	dir := t.TempDir()
	uc := newTestAssetCacheUC(t, dir, "", stubVRChatRunning{})
	n, err := uc.Clear()
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if n != 0 {
		t.Fatalf("n = %d, want 0", n)
	}
}

func TestClearVRChatAssetCache_missingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	uc := newTestAssetCacheUC(t, missing, "", stubVRChatRunning{})
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCachePathMissing) {
		t.Fatalf("err = %v, want ErrAssetCachePathMissing", err)
	}
}

func TestClearVRChatAssetCache_vrchatRunning(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.bin"), []byte("x"), 0o644)
	uc := newTestAssetCacheUC(t, dir, "", stubVRChatRunning{running: true})
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCacheVRChatRunning) {
		t.Fatalf("err = %v, want ErrAssetCacheVRChatRunning", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.bin")); err != nil {
		t.Fatalf("file should remain: %v", err)
	}
}

func TestClearVRChatAssetCache_rechecksRunningBeforeDelete(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.bin"), []byte("x"), 0o644)
	checks := 0
	uc := &VRChatAssetCacheUseCase{
		readConfig: func() (*vrchatconfig.VRChatConfig, error) {
			return &vrchatconfig.VRChatConfig{CacheDirectory: dir}, nil
		},
		running: stubVRChatRunningFn(func() (bool, error) {
			checks++
			return checks >= 2, nil // second check reports running
		}),
		defaultCache:   func() (string, error) { return dir, nil },
		defaultPicture: func() (string, error) { return filepath.Join(t.TempDir(), "pics"), nil },
		vrchatDataDir:  func() (string, error) { return filepath.Join(t.TempDir(), "vrc"), nil },
	}
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCacheVRChatRunning) {
		t.Fatalf("err = %v, want ErrAssetCacheVRChatRunning", err)
	}
	if checks < 2 {
		t.Fatalf("VRChatRunning checks = %d, want >= 2", checks)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.bin")); err != nil {
		t.Fatalf("file should remain when recheck fails: %v", err)
	}
}

type stubVRChatRunningFn func() (bool, error)

func (f stubVRChatRunningFn) VRChatRunning() (bool, error) { return f() }

func TestClearVRChatAssetCache_sameAsPictureFolder(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.bin"), []byte("x"), 0o644)
	uc := newTestAssetCacheUC(t, dir, dir, stubVRChatRunning{})
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCacheEqualsPictureFolder) {
		t.Fatalf("err = %v, want ErrAssetCacheEqualsPictureFolder", err)
	}
}

func TestClearVRChatAssetCache_volumeRoot(t *testing.T) {
	uc := &VRChatAssetCacheUseCase{
		readConfig: func() (*vrchatconfig.VRChatConfig, error) {
			root := "/"
			if runtime.GOOS == "windows" {
				root = filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
			}
			return &vrchatconfig.VRChatConfig{CacheDirectory: root}, nil
		},
		running:        stubVRChatRunning{},
		defaultCache:   func() (string, error) { return t.TempDir(), nil },
		defaultPicture: func() (string, error) { return filepath.Join(t.TempDir(), "pics"), nil },
		vrchatDataDir:  func() (string, error) { return filepath.Join(t.TempDir(), "vrc"), nil },
	}
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCacheVolumeRoot) {
		t.Fatalf("err = %v, want ErrAssetCacheVolumeRoot", err)
	}
}

func TestClearVRChatAssetCache_defaultPath(t *testing.T) {
	cache := t.TempDir()
	_ = os.WriteFile(filepath.Join(cache, "a.bin"), []byte("x"), 0o644)
	uc := &VRChatAssetCacheUseCase{
		readConfig: func() (*vrchatconfig.VRChatConfig, error) {
			return &vrchatconfig.VRChatConfig{CacheDirectory: ""}, nil
		},
		running:        stubVRChatRunning{},
		defaultCache:   func() (string, error) { return cache, nil },
		defaultPicture: func() (string, error) { return filepath.Join(t.TempDir(), "pics"), nil },
		vrchatDataDir:  func() (string, error) { return filepath.Join(t.TempDir(), "vrc"), nil },
	}
	n, err := uc.Clear()
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if n != 1 {
		t.Fatalf("n = %d, want 1", n)
	}
}

func TestClearVRChatAssetCache_partialFailure(t *testing.T) {
	dir := t.TempDir()
	okFile := filepath.Join(dir, "ok.bin")
	if err := os.WriteFile(okFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(blocked, "inner.bin")
	if err := os.WriteFile(inner, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make inner unreadable/undeletable on Unix; on Windows skip if chmod insufficient.
	if runtime.GOOS != "windows" {
		if err := os.Chmod(blocked, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(blocked, 0o755) })
	} else {
		t.Skip("partial delete lock is unreliable on Windows in CI")
	}

	uc := newTestAssetCacheUC(t, dir, "", stubVRChatRunning{})
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCacheRemoveFailed) {
		t.Fatalf("err = %v, want ErrAssetCacheRemoveFailed", err)
	}
	// Must not report success; ok.bin may or may not already be gone depending on ReadDir order.
}

func TestClearVRChatAssetCache_refusesVRChatDataDir(t *testing.T) {
	data := t.TempDir()
	_ = os.WriteFile(filepath.Join(data, "config.json"), []byte("{}"), 0o644)
	uc := &VRChatAssetCacheUseCase{
		readConfig: func() (*vrchatconfig.VRChatConfig, error) {
			return &vrchatconfig.VRChatConfig{CacheDirectory: data}, nil
		},
		running:        stubVRChatRunning{},
		defaultCache:   func() (string, error) { return filepath.Join(data, "Cache-WindowsPlayer"), nil },
		defaultPicture: func() (string, error) { return filepath.Join(t.TempDir(), "pics"), nil },
		vrchatDataDir:  func() (string, error) { return data, nil },
	}
	_, err := uc.Clear()
	if !errors.Is(err, ErrAssetCacheEqualsVRChatDataDir) {
		t.Fatalf("err = %v, want ErrAssetCacheEqualsVRChatDataDir", err)
	}
}

func TestClearVRChatAssetCache_symlinkNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation often needs elevation on Windows")
	}
	dir := t.TempDir()
	outside := t.TempDir()
	target := filepath.Join(outside, "keep.bin")
	if err := os.WriteFile(target, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link-out")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	uc := newTestAssetCacheUC(t, dir, "", stubVRChatRunning{})
	n, err := uc.Clear()
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if n != 1 {
		t.Fatalf("n = %d, want 1", n)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("symlink target must remain: %v", err)
	}
}

func TestResolveVRChatAssetCachePath_usesConfig(t *testing.T) {
	dir := t.TempDir()
	uc := newTestAssetCacheUC(t, dir, "", stubVRChatRunning{})
	got, err := uc.ResolvePath()
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Clean(dir) {
		t.Fatalf("got %q want %q", got, filepath.Clean(dir))
	}
}

func TestIsVolumeRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		if !isVolumeRoot(`C:\`) {
			t.Fatal("C:\\ should be volume root")
		}
		if isVolumeRoot(`C:\Users`) {
			t.Fatal("C:\\Users should not be volume root")
		}
	} else {
		if !isVolumeRoot("/") {
			t.Fatal("/ should be volume root")
		}
		if isVolumeRoot("/tmp") {
			t.Fatal("/tmp should not be volume root")
		}
	}
}

func TestSamePath(t *testing.T) {
	a := filepath.Join(t.TempDir(), "Foo")
	if !samePath(a, a) {
		t.Fatalf("samePath(%q, %q) = false", a, a)
	}
	if runtime.GOOS == "windows" {
		b := strings.ToUpper(a)
		if !samePath(a, b) {
			t.Fatalf("windows samePath(%q, %q) = false", a, b)
		}
	} else {
		other := filepath.Join(filepath.Dir(a), "foo")
		if samePath(a, other) && a != other {
			t.Fatalf("unix samePath should be case-sensitive: %q vs %q", a, other)
		}
	}
}

func newTestAssetCacheUC(t *testing.T, cacheDir, pictureDir string, running stubVRChatRunning) *VRChatAssetCacheUseCase {
	t.Helper()
	if pictureDir == "" {
		pictureDir = filepath.Join(t.TempDir(), "pictures-vrchat")
	}
	return &VRChatAssetCacheUseCase{
		readConfig: func() (*vrchatconfig.VRChatConfig, error) {
			return &vrchatconfig.VRChatConfig{CacheDirectory: cacheDir}, nil
		},
		running:        running,
		defaultCache:   func() (string, error) { return cacheDir, nil },
		defaultPicture: func() (string, error) { return pictureDir, nil },
		vrchatDataDir:  func() (string, error) { return filepath.Join(t.TempDir(), "vrchat-data"), nil },
	}
}
