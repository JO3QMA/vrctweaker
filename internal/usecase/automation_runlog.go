package usecase

import (
	"sync"
	"time"

	"vrchat-tweaker/internal/domain/automation"
)

type runLogStore struct {
	mu      sync.Mutex
	entries []automation.RunLogEntry
}

func newRunLogStore() *runLogStore {
	return &runLogStore{entries: make([]automation.RunLogEntry, 0, automation.RunLogMaxEntries)}
}

func (s *runLogStore) append(e automation.RunLogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
	if len(s.entries) > automation.RunLogMaxEntries {
		s.entries = s.entries[len(s.entries)-automation.RunLogMaxEntries:]
	}
}

func (s *runLogStore) list() []automation.RunLogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]automation.RunLogEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

type failureLogLimiter struct {
	mu    sync.Mutex
	last  map[string]time.Time
	limit time.Duration
}

func newFailureLogLimiter(limit time.Duration) *failureLogLimiter {
	return &failureLogLimiter{last: make(map[string]time.Time), limit: limit}
}

func (l *failureLogLimiter) allow(itemID string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t, ok := l.last[itemID]; ok && now.Sub(t) < l.limit {
		return false
	}
	l.last[itemID] = now
	return true
}
