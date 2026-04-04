package vrchatapi

import (
	"sync"
)

// credentialStoreWithFileFallback tries primary (e.g. keyring) first; if Set fails,
// saves to file and subsequent Get reads file after keyring miss. Covers Linux where
// a probe Set succeeds but a later Set hits dbus/keyring errors (or stale probe state).
type credentialStoreWithFileFallback struct {
	mu      sync.Mutex
	primary CredentialStore
	file    *FileCredentialStore
	warn    func(string)
}

func newCredentialStoreWithFileFallback(
	primary CredentialStore,
	file *FileCredentialStore,
	warn func(string),
) CredentialStore {
	if warn == nil {
		warn = func(string) {}
	}
	return &credentialStoreWithFileFallback{
		primary: primary,
		file:    file,
		warn:    warn,
	}
}

func (s *credentialStoreWithFileFallback) Get(service, user string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, err := s.primary.Get(service, user); err == nil && v != "" {
		return v, nil
	}
	return s.file.Get(service, user)
}

func (s *credentialStoreWithFileFallback) Set(service, user, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.primary.Set(service, user, password)
	if err == nil {
		if delErr := s.file.Delete(service, user); delErr != nil {
			s.warn("credential store: file fallback cleanup failed after keyring Set (" + delErr.Error() + ")")
		}
		return nil
	}
	s.warn("credential store: keyring Set failed; using file fallback (" + err.Error() + ")")
	return s.file.Set(service, user, password)
}

func (s *credentialStoreWithFileFallback) Delete(service, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.primary.Delete(service, user)
	return s.file.Delete(service, user)
}
