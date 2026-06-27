package logwatcher

import (
	"os"
	"sync"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

type stubParser struct {
	events []activity.ParsedEvent
	err    error
}

func (p stubParser) ParseLine(string, time.Time) ([]activity.ParsedEvent, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.events, nil
}

type raceSafeLogBuffer struct {
	mu   sync.Mutex
	logs []string
}

func (b *raceSafeLogBuffer) logger() Logger {
	return func(format string, args ...any) {
		b.mu.Lock()
		b.logs = append(b.logs, format)
		b.mu.Unlock()
	}
}

func (b *raceSafeLogBuffer) len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.logs)
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

func appendToTestFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write([]byte(content))
	return err
}
