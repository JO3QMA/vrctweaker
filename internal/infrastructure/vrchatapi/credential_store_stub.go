package vrchatapi

import (
	"errors"
	"sync"
)

// StubCredentialStore is an in-memory implementation for development.
// Replace with OS keyring (e.g., zalando/go-keyring) in production.
type StubCredentialStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewStubCredentialStore creates a new StubCredentialStore.
func NewStubCredentialStore() *StubCredentialStore {
	return &StubCredentialStore{data: make(map[string]string)}
}

// Get retrieves the stored value.
func (s *StubCredentialStore) Get(service, user string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := service + ":" + user
	if v, ok := s.data[key]; ok {
		return v, nil
	}
	return "", errors.New("not found")
}

// Set stores a value.
func (s *StubCredentialStore) Set(service, user, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[service+":"+user] = password
	return nil
}

// Delete removes a stored value.
func (s *StubCredentialStore) Delete(service, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, service+":"+user)
	return nil
}
