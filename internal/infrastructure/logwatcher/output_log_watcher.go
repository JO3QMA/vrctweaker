package logwatcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"sync"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/diag"
)

// Logger is a diagnostic logger for logwatcher (shared with picturewatcher via diag).
type Logger = diag.Logger

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

// LogFileSwitchHandler is invoked when the watcher begins tailing a different output_log file,
// or when the current file was truncated in place. previousPath is empty on the first file.
type LogFileSwitchHandler interface {
	OnLogFileSwitch(ctx context.Context, previousPath, newPath string) error
}

// LogFileSwitchHandlerFunc adapts a function to LogFileSwitchHandler.
type LogFileSwitchHandlerFunc func(ctx context.Context, previousPath, newPath string) error

// OnLogFileSwitch implements LogFileSwitchHandler.
func (f LogFileSwitchHandlerFunc) OnLogFileSwitch(ctx context.Context, previousPath, newPath string) error {
	if f == nil {
		return nil
	}
	return f(ctx, previousPath, newPath)
}

// LogParser parses a log line into events.
type LogParser interface {
	ParseLine(line string, baseTime time.Time) ([]activity.ParsedEvent, error)
}

// OutputLogWatcher tails output_log.txt and emits parsed events.
// configuredPath may be a regular file or a directory; if it is a directory, the newest
// output_log*.txt under it is tailed and the watcher switches when a newer file appears.
type OutputLogWatcher struct {
	configuredPath string
	watchDir       string // non-empty => resolve latest output_log*.txt under this dir
	fixedFile      string // non-empty => tail this file only
	parser         LogParser
	handler        EventHandler
	logger         Logger
	// logFileSwitchHandler is optional; called when tailing switches to another output_log file
	// or when the current file was truncated. Used to finalize open activity rows and ingest
	// startup lines already written to the new file before the watcher seeks to EOF.
	logFileSwitchHandler LogFileSwitchHandler

	mu        sync.Mutex
	status    string // "idle", "running", "stopped"
	lastErr   error
	lastErrAt time.Time

	// lastTailedPath is written only from run(); tracks the path last opened for tailing.
	lastTailedPath string
}

// NewOutputLogWatcher creates a watcher for the given path (file or directory).
func NewOutputLogWatcher(configuredPath string, parser LogParser, handler EventHandler, logger Logger) *OutputLogWatcher {
	if logger == nil {
		logger = diag.Nop
	}
	w := &OutputLogWatcher{
		configuredPath: configuredPath,
		parser:         parser,
		handler:        handler,
		logger:         logger,
		status:         "idle",
	}
	if info, err := os.Stat(configuredPath); err == nil && info.IsDir() {
		w.watchDir = configuredPath
	} else {
		w.fixedFile = configuredPath
	}
	return w
}

// SetLogFileSwitchHandler registers a callback invoked when the tailed output_log path changes
// or the current file is truncated. Call before Start.
func (w *OutputLogWatcher) SetLogFileSwitchHandler(h LogFileSwitchHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.logFileSwitchHandler = h
}

func (w *OutputLogWatcher) resolveActivePath() (string, error) {
	if w.watchDir != "" {
		return ResolveLatestOutputLogFile(w.watchDir)
	}
	if w.fixedFile != "" {
		return w.fixedFile, nil
	}
	return "", os.ErrInvalid
}

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

		activePath, resolveErr := w.resolveActivePath()
		if resolveErr != nil {
			w.setErr(resolveErr)
			w.logger("[logwatcher] resolve active log: %v", resolveErr)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reopenBackoff):
				continue
			}
		}

		f, err := os.Open(activePath)
		if err != nil {
			w.setErr(err)
			w.logger("[logwatcher] open %s: %v", activePath, err)
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
			w.logger("[logwatcher] stat: %v", err)
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

		if w.lastTailedPath != "" && activePath != w.lastTailedPath {
			w.mu.Lock()
			h := w.logFileSwitchHandler
			w.mu.Unlock()
			if h != nil {
				if switchErr := h.OnLogFileSwitch(ctx, w.lastTailedPath, activePath); switchErr != nil {
					w.logger("[logwatcher] log file switch: %v", switchErr)
				}
			}
		}
		w.lastTailedPath = activePath

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
					w.logger("[logwatcher] read error: %v", err)
					break readLoop
				}
				// EOF: check for file rotation (truncate or replace) or newer log file in dir mode
				curInfo, statErr := os.Stat(activePath)
				if statErr != nil {
					_ = f.Close()
					break readLoop
				}
				if curInfo.Size() < initialSize {
					w.mu.Lock()
					h := w.logFileSwitchHandler
					w.mu.Unlock()
					if h != nil {
						if switchErr := h.OnLogFileSwitch(ctx, activePath, activePath); switchErr != nil {
							w.logger("[logwatcher] log file truncate: %v", switchErr)
						}
					}
					_ = f.Close()
					break readLoop
				}
				if curInfo.ModTime() != info.ModTime() {
					_ = f.Close()
					break readLoop
				}
				if w.watchDir != "" {
					latest, latErr := ResolveLatestOutputLogFile(w.watchDir)
					if latErr == nil && latest != activePath {
						_ = f.Close()
						w.logger("[logwatcher] switching to newer output log: %s", latest)
						break readLoop
					}
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

			baseTime := activity.ParseVRChatTimestamp(lineTrimmed, time.Now().In(time.Local))
			events, parseErr := w.parser.ParseLine(lineTrimmed, baseTime)
			if parseErr != nil {
				w.logger("[logwatcher] parse error: %v", parseErr)
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
