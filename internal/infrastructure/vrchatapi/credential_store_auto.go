package vrchatapi

import (
	"os"
	"path/filepath"
	"runtime"
)

// NewAutoCredentialStore uses the OS keyring when it works; on Linux, if a probe
// Set/Delete fails (no D-Bus / secret service), falls back to a file in dataDir.
// When the probe passes but a later Set still fails (e.g. dbus-launch), Set falls back
// to the same file via credentialStoreWithFileFallback.
// Set VRCHAT_TWEAKER_USE_FILE_CREDENTIALS=1 to force the file store only.
func NewAutoCredentialStore(dataDir string, warn func(string)) CredentialStore {
	filePath := filepath.Join(dataDir, ".vrchat-auth-token")
	fileStore := NewFileCredentialStore(filePath)
	if os.Getenv("VRCHAT_TWEAKER_USE_FILE_CREDENTIALS") != "" {
		return fileStore
	}
	k := NewKeyringCredentialStore()
	if runtime.GOOS != "linux" {
		return newCredentialStoreWithFileFallback(k, fileStore, warn)
	}
	if err := probeKeyringWritable(k); err != nil {
		if warn != nil {
			warn("keyring unavailable; using file-backed credentials (" + err.Error() + ")")
		}
		return fileStore
	}
	return newCredentialStoreWithFileFallback(k, fileStore, warn)
}

func probeKeyringWritable(k *KeyringCredentialStore) error {
	u := CredentialUser + "-probe"
	_ = k.Delete(CredentialService, u)
	if err := k.Set(CredentialService, u, "x"); err != nil {
		return err
	}
	return k.Delete(CredentialService, u)
}
