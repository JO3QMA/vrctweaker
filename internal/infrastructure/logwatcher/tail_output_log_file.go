package logwatcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"
)

const (
	pollInterval   = 500 * time.Millisecond
	readBufferSize = 64 * 1024
)

// tailCheckpoint reports read progress after each consumed line.
type tailCheckpoint func(byteOffset int64, vrChatLineTime time.Time)

// tailOutputLogFile reads new lines from path starting at startOffset until ctx is cancelled.
func tailOutputLogFile(
	ctx context.Context,
	path string,
	startOffset int64,
	parser LogParser,
	handler EventHandler,
	logger Logger,
	checkpoint tailCheckpoint,
) {
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

	br := bufio.NewReaderSize(f, readBufferSize)
	offset := startOffset

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := br.ReadString('\n')
		consumed := int64(len(line))
		if consumed > 0 {
			offset += consumed
		}

		if err != nil {
			if err != io.EOF {
				logger("[multi-logwatcher] read %s: %v", path, err)
				return
			}
			if consumed == 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(pollInterval):
				}
				continue
			}
		}

		lineTrimmed := trimNL(line)
		if lineTrimmed != "" {
			baseTime, parseErr := dispatchOutputLogLine(lineTrimmed, parser, handler)
			if parseErr != nil {
				logDispatchLineErr(logger, parseErr,
					"[multi-logwatcher] parse %s: %v", "[multi-logwatcher] dispatch %s: %v",
					path)
			}
			if checkpoint != nil {
				checkpoint(offset, baseTime)
			}
		}

		if err == io.EOF {
			select {
			case <-ctx.Done():
				return
			case <-time.After(pollInterval):
			}
		}
	}
}
