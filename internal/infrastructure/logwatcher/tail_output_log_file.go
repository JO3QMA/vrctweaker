package logwatcher

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

const (
	pollInterval   = 500 * time.Millisecond
	readBufferSize = 64 * 1024

	tailIngestRetryInitial = 500 * time.Millisecond
	tailIngestRetryMax     = 30 * time.Second
)

// tailIngestRetryWait sleeps before retrying a failed live-tail ingest. Tests may replace the func under tailIngestRetryWaitMu.
var (
	tailIngestRetryWaitMu sync.RWMutex
	tailIngestRetryWait   = defaultTailIngestRetryWait
)

func defaultTailIngestRetryWait(ctx context.Context, attempt int) bool {
	delay := tailIngestRetryDelay(attempt)
	select {
	case <-ctx.Done():
		return true
	case <-time.After(delay):
		return false
	}
}

func waitTailIngestRetry(ctx context.Context, attempt int) bool {
	tailIngestRetryWaitMu.RLock()
	fn := tailIngestRetryWait
	tailIngestRetryWaitMu.RUnlock()
	return fn(ctx, attempt)
}

func tailIngestRetryDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := tailIngestRetryInitial
	for i := 0; i < attempt && delay < tailIngestRetryMax; i++ {
		delay *= 2
		if delay > tailIngestRetryMax {
			delay = tailIngestRetryMax
			break
		}
	}
	if delay > tailIngestRetryMax {
		delay = tailIngestRetryMax
	}
	return delay
}

// tailCheckpoint reports read progress after each consumed line.
type tailCheckpoint func(byteOffset int64, vrChatLineTime time.Time)

type tailCallbacks struct {
	checkpoint        tailCheckpoint
	setIngestRetrying func(active bool)
	onIngestRetryWait func()
}

// tailOutputLogFile reads new lines from path starting at startOffset until ctx is cancelled.
func tailOutputLogFile(
	ctx context.Context,
	path string,
	startOffset int64,
	parser *activity.LogParser,
	handler EventHandler,
	logger Logger,
	cb tailCallbacks,
) {
	if logger == nil {
		logger = Nop
	}
	f, err := os.Open(path)
	if err != nil {
		logger("[multi-logwatcher] open %s: %v", path, err)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
		logger("[multi-logwatcher] seek %s: %v", path, err)
		return
	}

	processor := NewLineProcessor(parser, handler)
	br := bufio.NewReaderSize(f, readBufferSize)
	offset := startOffset

	// Intentionally not using scanOutputLogFromReader: tail must poll on EOF and on
	// zero-byte reads (waitForTail) until ctx is cancelled, which batch scan does not do.
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		lineStartOffset := offset
		line, err := br.ReadString('\n')
		consumed := int64(len(line))
		lineEndOffset := lineStartOffset + consumed
		if consumed > 0 {
			offset = lineEndOffset
		}

		if err != nil {
			if err != io.EOF {
				logger("[multi-logwatcher] read %s: %v", path, err)
				return
			}
			if consumed == 0 {
				if waitForTail(ctx) {
					return
				}
				continue
			}
		}

		lineTrimmed := trimNL(line)
		if lineTrimmed != "" {
			attempt := 0
			for {
				baseTime, parseErr := processor.Process(lineTrimmed)
				if parseErr == nil {
					if cb.checkpoint != nil {
						cb.checkpoint(lineEndOffset, baseTime)
					}
					break
				}
				if !errors.Is(parseErr, ErrHandlerDispatch) {
					logLineProcessErr(logger, parseErr,
						"[multi-logwatcher] parse %s: %v", "[multi-logwatcher] dispatch %s: %v",
						path)
					break
				}
				delay := tailIngestRetryDelay(attempt)
				logger("[multi-logwatcher] ingest %s at offset %d failed (attempt %d): %v; retry in %s",
					path, lineStartOffset, attempt+1, parseErr, delay)
				if cb.setIngestRetrying != nil {
					cb.setIngestRetrying(true)
				}
				if cb.onIngestRetryWait != nil {
					cb.onIngestRetryWait()
				}
				if waitTailIngestRetry(ctx, attempt) {
					if cb.setIngestRetrying != nil {
						cb.setIngestRetrying(false)
					}
					return
				}
				attempt++
				offset = lineStartOffset
				if _, seekErr := f.Seek(lineStartOffset, io.SeekStart); seekErr != nil {
					logger("[multi-logwatcher] seek %s for ingest retry: %v", path, seekErr)
					return
				}
				br.Reset(f)
				line, err = br.ReadString('\n')
				consumed = int64(len(line))
				lineEndOffset = lineStartOffset + consumed
				if consumed > 0 {
					offset = lineEndOffset
				}
				if err != nil && err != io.EOF {
					logger("[multi-logwatcher] read %s on ingest retry: %v", path, err)
					return
				}
				lineTrimmed = trimNL(line)
				if lineTrimmed == "" {
					logger("[multi-logwatcher] ingest retry %s at offset %d: line became empty", path, lineStartOffset)
					break
				}
			}
			if cb.setIngestRetrying != nil {
				cb.setIngestRetrying(false)
			}
		}

		if err == io.EOF {
			if waitForTail(ctx) {
				return
			}
		}
	}
}

func waitForTail(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(pollInterval):
		return false
	}
}
