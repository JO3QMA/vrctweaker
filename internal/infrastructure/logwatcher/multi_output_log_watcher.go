package logwatcher

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	"vrchat-tweaker/internal/infrastructure/diag"
)

const defaultActiveLogStallTimeout = 60 * time.Second

// MultiOutputLogWatcherCallbacks receives lifecycle events for multi-source tailing.
type MultiOutputLogWatcherCallbacks struct {
	OnLogRotationHandoff func(ctx context.Context, oldPath string) error
	OnTailCheckpoint     func(ctx context.Context, path string, byteOffset int64, vrChatLineTime time.Time)
}

// MultiOutputLogWatcher polls a log directory and tails every output_log*.txt that is growing.
type MultiOutputLogWatcher struct {
	watchDir       string
	parser         LogParser
	handlerFactory func(logPath string) EventHandler
	callbacks      MultiOutputLogWatcherCallbacks
	logger         Logger

	activeLogStallTimeout time.Duration

	*pollWatcherState
}

type trackedLogFile struct {
	lastSize     int64
	lastGrowthAt time.Time
	readOffset   atomic.Int64
	tailing      atomic.Bool
	tailGen      atomic.Uint64
	cancel       context.CancelFunc
}

// NewMultiOutputLogWatcher creates a directory-only watcher for parallel output_log sources.
func NewMultiOutputLogWatcher(
	watchDir string,
	parser LogParser,
	handlerFactory func(logPath string) EventHandler,
	callbacks MultiOutputLogWatcherCallbacks,
	logger Logger,
) *MultiOutputLogWatcher {
	if logger == nil {
		logger = diag.Nop
	}
	return &MultiOutputLogWatcher{
		watchDir:              watchDir,
		parser:                parser,
		handlerFactory:        handlerFactory,
		callbacks:             callbacks,
		logger:                logger,
		activeLogStallTimeout: defaultActiveLogStallTimeout,
		pollWatcherState:      newPollWatcherState(),
	}
}

// Start begins polling in a goroutine. Cancel ctx to stop.
func (w *MultiOutputLogWatcher) Start(ctx context.Context) error {
	if !w.tryStart() {
		return nil
	}
	go w.run(ctx)
	return nil
}

func (w *MultiOutputLogWatcher) run(ctx context.Context) {
	defer w.setStatus(statusStopped)

	tracked := make(map[string]*trackedLogFile)
	now := time.Now()
	baselineDone := false

	for {
		select {
		case <-ctx.Done():
			w.stopAllTails(tracked)
			return
		default:
		}

		files, err := ListOutputLogFiles(w.watchDir)
		if err != nil {
			if err != ErrNoOutputLogFiles {
				w.setErr(err)
				w.logger("[multi-logwatcher] list logs: %v", err)
			}
			select {
			case <-ctx.Done():
				w.stopAllTails(tracked)
				return
			case <-time.After(pollInterval):
				now = time.Now()
			}
			continue
		}
		w.setErr(nil)

		if !baselineDone {
			for _, path := range files {
				info, statErr := os.Stat(path)
				if statErr != nil {
					continue
				}
				tracked[path] = &trackedLogFile{lastSize: info.Size()}
			}
			baselineDone = true
			select {
			case <-ctx.Done():
				w.stopAllTails(tracked)
				return
			case <-time.After(pollInterval):
				now = time.Now()
			}
			continue
		}

		growingNow := make(map[string]int64)
		currentSizes := make(map[string]int64, len(files))

		for _, path := range files {
			info, statErr := os.Stat(path)
			if statErr != nil {
				continue
			}
			size := info.Size()
			currentSizes[path] = size

			state, ok := tracked[path]
			if !ok {
				state = &trackedLogFile{}
				tracked[path] = state
			}

			if size > state.lastSize {
				growingNow[path] = state.lastSize
				state.lastGrowthAt = now
				w.startTail(ctx, path, state, state.lastSize)
				state.lastSize = size
			} else if size < state.lastSize {
				// Truncated or rotated in place; restart from current size.
				w.stopTail(state)
				state.readOffset.Store(size)
				state.lastSize = size
				state.lastGrowthAt = now
			} else {
				state.lastSize = size
			}
		}

		for path, state := range tracked {
			if _, listed := currentSizes[path]; !listed {
				w.stopTail(state)
				delete(tracked, path)
				continue
			}
			if _, growing := growingNow[path]; growing {
				continue
			}
			if !state.tailing.Load() {
				continue
			}
			if w.shouldHandoff(state, now, growingNow, path) {
				w.stopTail(state)
				if w.callbacks.OnLogRotationHandoff != nil {
					if handoffErr := w.callbacks.OnLogRotationHandoff(ctx, path); handoffErr != nil {
						w.logger("[multi-logwatcher] rotation handoff %s: %v", path, handoffErr)
					}
				}
				continue
			}
			if now.Sub(state.lastGrowthAt) >= w.activeLogStallTimeout {
				w.stopTail(state)
			}
		}

		select {
		case <-ctx.Done():
			w.stopAllTails(tracked)
			return
		case <-time.After(pollInterval):
			now = time.Now()
		}
	}
}

func (w *MultiOutputLogWatcher) shouldHandoff(
	state *trackedLogFile,
	now time.Time,
	growingNow map[string]int64,
	path string,
) bool {
	if now.Sub(state.lastGrowthAt) < pollInterval {
		return false
	}
	for otherPath := range growingNow {
		if otherPath != path {
			return true
		}
	}
	return false
}

func (w *MultiOutputLogWatcher) startTail(ctx context.Context, path string, state *trackedLogFile, startOffset int64) {
	if state.tailing.Load() {
		return
	}
	state.readOffset.Store(startOffset)
	tailCtx, cancel := context.WithCancel(ctx)
	myGen := state.tailGen.Add(1)
	state.cancel = cancel
	state.tailing.Store(true)

	handler := w.handlerFactory(path)
	if handler == nil {
		state.tailing.Store(false)
		state.cancel = nil
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				w.logger("[multi-logwatcher] tail panic %s: %v", path, r)
			}
			if state.tailGen.Load() == myGen {
				state.tailing.Store(false)
				state.cancel = nil
			}
		}()
		tailOutputLogFile(tailCtx, path, startOffset, w.parser, handler, w.logger, func(offset int64, lineTime time.Time) {
			state.readOffset.Store(offset)
			if w.callbacks.OnTailCheckpoint != nil {
				w.callbacks.OnTailCheckpoint(ctx, path, offset, lineTime)
			}
		})
	}()
}

func (w *MultiOutputLogWatcher) stopTail(state *trackedLogFile) {
	if !state.tailing.Load() {
		return
	}
	state.tailGen.Add(1)
	if state.cancel != nil {
		state.cancel()
		state.cancel = nil
	}
	state.tailing.Store(false)
}

func (w *MultiOutputLogWatcher) stopAllTails(tracked map[string]*trackedLogFile) {
	for _, state := range tracked {
		w.stopTail(state)
	}
}
