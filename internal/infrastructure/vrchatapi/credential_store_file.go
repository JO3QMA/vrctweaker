package vrchatapi

import (
	"errors"
	"os"
	"strings"
)

// FileCredentialStore persists the VRChat auth token in a single file (0600).
// Used when the OS keyring is unavailable (e.g. headless Linux / devcontainer without D-Bus).
type FileCredentialStore struct {
	path string
}

// NewFileCredentialStore creates a store that reads/writes path.
func NewFileCredentialStore(path string) *FileCredentialStore {
	return &FileCredentialStore{path: path}
}

// Get retrieves the stored auth token.
func (f *FileCredentialStore) Get(service, user string) (string, error) {
	if service != CredentialService || user != CredentialUser {
		return "", errors.New("not found")
	}
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("not found")
		}
		return "", err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return "", errors.New("not found")
	}
	return s, nil
}

// Set stores the auth token.
func (f *FileCredentialStore) Set(service, user, password string) error {
	if service != CredentialService || user != CredentialUser {
		return errors.New("unsupported credential key")
	}
	return os.WriteFile(f.path, []byte(password), 0600)
}

// Delete removes the stored token file.
func (f *FileCredentialStore) Delete(service, user string) error {
	if service != CredentialService || user != CredentialUser {
		return nil
	}
	if err := os.Remove(f.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
