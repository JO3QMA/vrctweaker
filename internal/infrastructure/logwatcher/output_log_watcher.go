package logwatcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"sync"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// Logger is a minimal interface for watcher logging.
type Logger interface {
	Printf(format string, args ...interface{})
}

// EventHandler receives parsed events from the watcher.
type EventHandler interface {
	Handle(event activity.ParsedEvent)
}

// EventHandlerFunc adapts a function to EventHandler.
type EventHandlerFunc func(event activity.ParsedEvent)

// Handle implements EventHandler.
func (f EventHandlerFunc) Handle(event activity.ParsedEvent) {
	f(event)
}

// LogParser parses a log line into events.
type LogParser interface {
	ParseLine(line string, baseTime time.Time) ([]activity.ParsedEvent, error)
}

// OutputLogWatcher tails output_log.txt and emits parsed events.
type OutputLogWatcher struct {
	path    string
	parser  LogParser
	handler EventHandler
	logger  Logger

	mu        sync.Mutex
	status    string // "idle", "running", "stopped"
	lastErr   error
	lastErrAt time.Time
}

// NewOutputLogWatcher creates a watcher for the given path.
func NewOutputLogWatcher(path string, parser LogParser, handler EventHandler, logger Logger) *OutputLogWatcher {
	if logger == nil {
		logger = nopLogger{}
	}
	return &OutputLogWatcher{
		path:    path,
		parser:  parser,
		handler: handler,
		logger:  logger,
		status:  "idle",
	}
}

type nopLogger struct{}

func (nopLogger) Printf(format string, args ...interface{}) {}

// Start begins tailing the file in a goroutine. Cancel ctx to stop.
func (w *OutputLogWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.status == "running" {
		w.mu.Unlock()
		return nil
	}
	w.status = "running"
	w.lastErr = nil
	w.mu.Unlock()

	go w.run(ctx)
	return nil
}

// Status returns the current status and last error if any.
func (w *OutputLogWatcher) Status() (status string, lastErr error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status, w.lastErr
}

func (w *OutputLogWatcher) setErr(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastErr = err
	w.lastErrAt = time.Now()
}

func (w *OutputLogWatcher) setStatus(s string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = s
}

const (
	pollInterval   = 500 * time.Millisecond
	reopenBackoff  = 2 * time.Second
	readBufferSize = 64 * 1024
)

func (w *OutputLogWatcher) run(ctx context.Context) {
	defer w.setStatus("stopped")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		f, err := os.Open(w.path)
		if err != nil {
			w.setErr(err)
			w.logger.Printf("[logwatcher] open %s: %v", w.path, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reopenBackoff):
				continue
			}
		}

		info, err := f.Stat()
		if err != nil {
			_ = f.Close()
			w.setErr(err)
			w.logger.Printf("[logwatcher] stat: %v", err)
			time.Sleep(reopenBackoff)
			continue
		}
		initialSize := info.Size()
		// Seek to end to tail only new content
		if _, err := f.Seek(initialSize, io.SeekStart); err != nil {
			_ = f.Close()
			w.setErr(err)
			time.Sleep(reopenBackoff)
			continue
		}

		w.setErr(nil)
		br := bufio.NewReaderSize(f, readBufferSize)

	readLoop:
		for {
			select {
			case <-ctx.Done():
				break readLoop
			default:
			}

			line, err := br.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					_ = f.Close()
					w.setErr(err)
					w.logger.Printf("[logwatcher] read error: %v", err)
					break readLoop
				}
				// EOF: check for file rotation (truncate or replace)
				curInfo, statErr := os.Stat(w.path)
				if statErr != nil {
					_ = f.Close()
					break readLoop
				}
				if curInfo.ModTime() != info.ModTime() || curInfo.Size() < initialSize {
					_ = f.Close()
					break readLoop
				}
				select {
				case <-ctx.Done():
					break readLoop
				case <-time.After(pollInterval):
				}
				continue
			}

			lineTrimmed := trimNL(line)
			if lineTrimmed == "" {
				continue
			}

			events, parseErr := w.parser.ParseLine(lineTrimmed, time.Now().UTC())
			if parseErr != nil {
				w.logger.Printf("[logwatcher] parse error: %v", parseErr)
				continue
			}
			for _, ev := range events {
				if ev != nil {
					w.handler.Handle(ev)
				}
			}
		}

		_ = f.Close()
		// Brief pause before reopening (e.g. after rotation)
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
