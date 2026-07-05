package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

func TestVRChatConfigFileRepository_Exists_NoFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	repo := NewVRChatConfigFileRepository(path)

	exists, err := repo.Exists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for non-existent file")
	}
}

func TestVRChatConfigFileRepository_WriteAndRead(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	repo := NewVRChatConfigFileRepository(path)

	boolTrue := true
	cfg := &vrchatconfig.VRChatConfig{
		CameraResWidth:           1920,
		CameraResHeight:          1080,
		ScreenshotResWidth:       3840,
		ScreenshotResHeight:      2160,
		CacheSize:                50,
		CacheExpiryDelay:         60,
		DisableRichPresence:      &boolTrue,
		PictureOutputSplitByDate: &boolTrue,
	}

	if err := repo.Write(cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	exists, err := repo.Exists()
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Error("expected file to exist after write")
	}

	got, err := repo.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.CameraResWidth != 1920 {
		t.Errorf("CameraResWidth: got %d, want 1920", got.CameraResWidth)
	}
	if got.CameraResHeight != 1080 {
		t.Errorf("CameraResHeight: got %d, want 1080", got.CameraResHeight)
	}
	if got.ScreenshotResWidth != 3840 {
		t.Errorf("ScreenshotResWidth: got %d, want 3840", got.ScreenshotResWidth)
	}
	if got.ScreenshotResHeight != 2160 {
		t.Errorf("ScreenshotResHeight: got %d, want 2160", got.ScreenshotResHeight)
	}
	if got.CacheSize != 50 {
		t.Errorf("CacheSize: got %d, want 50", got.CacheSize)
	}
	if got.DisableRichPresence == nil || !*got.DisableRichPresence {
		t.Error("DisableRichPresence: expected true")
	}
}

func TestVRChatConfigFileRepository_Delete(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	repo := NewVRChatConfigFileRepository(path)

	if err := repo.Write(&vrchatconfig.VRChatConfig{}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if err := repo.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	exists, err := repo.Exists()
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Error("expected file to not exist after delete")
	}
}

func TestVRChatConfigFileRepository_Delete_NonExistent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nonexistent.json")
	repo := NewVRChatConfigFileRepository(path)

	if err := repo.Delete(); err != nil {
		t.Fatalf("Delete non-existent should not error: %v", err)
	}
}

func TestVRChatConfigFileRepository_Read_NonExistent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "missing.json")
	repo := NewVRChatConfigFileRepository(path)

	_, err := repo.Read()
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestVRChatConfigFileRepository_Read_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	repo := NewVRChatConfigFileRepository(path)
	_, err := repo.Read()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestVRChatConfigFileRepository_Write_CreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "subdir", "config.json")
	repo := NewVRChatConfigFileRepository(path)

	if err := repo.Write(&vrchatconfig.VRChatConfig{CacheSize: 30}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := repo.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.CacheSize != 30 {
		t.Errorf("CacheSize: got %d, want 30", got.CacheSize)
	}
}

func TestVRChatConfigFileRepository_OmitsZeroValues(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	repo := NewVRChatConfigFileRepository(path)

	cfg := &vrchatconfig.VRChatConfig{
		CameraResWidth:  1920,
		CameraResHeight: 1080,
	}
	if err := repo.Write(cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if contains(content, "cache_size") {
		t.Error("expected zero-value fields to be omitted")
	}
	if !contains(content, "camera_res_width") {
		t.Error("expected camera_res_width to be present")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestVRChatConfigFileRepository_Write_ReadOnlyParent(t *testing.T) {
	tmp := t.TempDir()
	ro := filepath.Join(tmp, "readonly")
	if err := os.Mkdir(ro, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0755) })

	path := filepath.Join(ro, "nested", "config.json")
	repo := NewVRChatConfigFileRepository(path)
	err := repo.Write(&vrchatconfig.VRChatConfig{CacheSize: 1})
	if err == nil {
		t.Fatal("expected error writing under read-only directory")
	}
}

func TestVRChatConfigFileRepository_Read_DirectoryPath(t *testing.T) {
	tmp := t.TempDir()
	repo := NewVRChatConfigFileRepository(tmp)
	_, err := repo.Read()
	if err == nil {
		t.Fatal("expected error reading directory")
	}
}

func TestVRChatConfigFileRepository_Exists_NonExistentNested(t *testing.T) {
	repo := NewVRChatConfigFileRepository(filepath.Join(t.TempDir(), "a", "b", "config.json"))
	exists, err := repo.Exists()
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatal("expected false")
	}
}

func TestVRChatConfigFileRepository_Write_FilePathIsDirectory(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "blocked"), 0555); err != nil {
		t.Fatal(err)
	}
	repo := NewVRChatConfigFileRepository(filepath.Join(tmp, "blocked"))
	err := repo.Write(&vrchatconfig.VRChatConfig{CacheSize: 5})
	if err == nil {
		t.Fatal("expected WriteFile error when path is a directory")
	}
}

func TestVRChatConfigFileRepository_Delete_permissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	if err := os.WriteFile(path, []byte("{}"), 0400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(tmp, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(tmp, 0755) })
	repo := NewVRChatConfigFileRepository(path)
	if err := repo.Delete(); err == nil {
		t.Fatal("expected delete error in read-only directory")
	}
}

func TestVRChatConfigFileRepository_Exists_permissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "secret")
	if err := os.Mkdir(sub, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(sub, 0700) })
	repo := NewVRChatConfigFileRepository(filepath.Join(sub, "config.json"))
	_, err := repo.Exists()
	if err == nil {
		t.Fatal("expected stat permission error")
	}
}
