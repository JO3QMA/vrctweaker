package logwatcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// ProcessOutputLogFile reads an entire output_log file from the beginning and dispatches parsed events.
func ProcessOutputLogFile(ctx context.Context, path string, parser LogParser, handler EventHandler, logger Logger) error {
	_, err := ProcessOutputLogFileFromOffset(ctx, path, 0, parser, handler, logger, nil)
	return err
}

// ProgressCallback receives absolute byte offset in the file after each line (including newline) and the raw line text.
type ProgressCallback func(byteOffset int64, line string)

// ProcessOutputLogFileFromOffset reads from startOffset and returns the final byte offset in the file.
func ProcessOutputLogFileFromOffset(ctx context.Context, path string, startOffset int64, parser LogParser, handler EventHandler, logger Logger, onProgress ProgressCallback) (int64, error) {
	if logger == nil {
		logger = nopLogger{}
	}
	f, err := os.Open(path)
	if err != nil {
		return startOffset, err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return startOffset, err
	}
	size := info.Size()
	if startOffset < 0 {
		startOffset = 0
	}
	if startOffset > size {
		startOffset = size
	}
	if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
		return startOffset, err
	}

	br := bufio.NewReader(f)
	pos := startOffset
	for {
		select {
		case <-ctx.Done():
			return pos, ctx.Err()
		default:
		}
		lineBytes, err := br.ReadBytes('\n')
		if len(lineBytes) == 0 && err == io.EOF {
			break
		}
		pos += int64(len(lineBytes))
		line := string(lineBytes)
		if err != nil && err != io.EOF {
			return pos, err
		}
		lineTrimmed := trimNL(line)
		if lineTrimmed != "" {
			baseTime := activity.ParseVRChatTimestamp(lineTrimmed, time.Now().In(time.Local))
			events, parseErr := parser.ParseLine(lineTrimmed, baseTime)
			if parseErr != nil {
				logger.Printf("[logwatcher] bootstrap parse error: %v", parseErr)
			} else {
				for _, ev := range events {
					if ev != nil {
						handler.Handle(ev)
					}
				}
			}
		}
		if onProgress != nil {
			onProgress(pos, trimNL(line))
		}
		if err == io.EOF {
			break
		}
	}
	return pos, nil
}
