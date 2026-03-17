package vrchatapi

import (
	"github.com/zalando/go-keyring"
)

// KeyringCredentialStore stores auth tokens in OS secure storage
// (Windows Credential Manager, macOS Keychain, Linux Secret Service).
type KeyringCredentialStore struct{}

// NewKeyringCredentialStore creates a new KeyringCredentialStore.
func NewKeyringCredentialStore() *KeyringCredentialStore {
	return &KeyringCredentialStore{}
}

// Get retrieves the stored auth token.
func (k *KeyringCredentialStore) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

// Set stores the auth token.
func (k *KeyringCredentialStore) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

// Delete removes the stored auth token.
func (k *KeyringCredentialStore) Delete(service, user string) error {
	return keyring.Delete(service, user)
}
