package logwatcher

import "sync"

type watcherStatus string

const (
	statusIdle    watcherStatus = "idle"
	statusRunning watcherStatus = "running"
	statusStopped watcherStatus = "stopped"
)

// pollWatcherState holds status and last error for tail/poll log watchers.
type pollWatcherState struct {
	mu      sync.Mutex
	status  watcherStatus
	lastErr error
}

func newPollWatcherState() *pollWatcherState {
	return &pollWatcherState{status: statusIdle}
}

func (s *pollWatcherState) Status() (status string, lastErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s.status), s.lastErr
}

func (s *pollWatcherState) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastErr = err
}

func (s *pollWatcherState) setStopped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != statusRunning {
		return
	}
	s.status = statusStopped
}

// tryStart marks the watcher running unless already running.
// Restart after statusStopped is intentional: ctx cancel stops the goroutine, then Start may run again.
func (s *pollWatcherState) tryStart() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == statusRunning {
		return false
	}
	s.status = statusRunning
	s.lastErr = nil
	return true
}
