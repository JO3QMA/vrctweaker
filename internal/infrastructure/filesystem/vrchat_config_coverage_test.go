package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

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
