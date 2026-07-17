package logwatcher

import (
	"bufio"
	"context"
	"io"
	"os"

	"vrchat-tweaker/internal/domain/activity"
)

// SessionCorrelatorWarmer rebuilds SessionCorrelator state from log lines without persisting commands.
type SessionCorrelatorWarmer interface {
	WarmFromParsedEvent(event activity.ParsedEvent)
}

// ProcessOutputLogFile reads an entire output_log file from the beginning and dispatches parsed events.
func ProcessOutputLogFile(ctx context.Context, path string, parser *activity.LogParser, handler EventHandler, logger Logger) error {
	_, err := ProcessOutputLogFileFromOffset(ctx, path, 0, parser, handler, logger, nil)
	return err
}

// WarmSessionCorrelatorFromLogFile replays log lines in [0, endOffset) into correlator only.
// Used before bootstrap resume so mid-file checkpoints keep correct world/instance context.
func WarmSessionCorrelatorFromLogFile(ctx context.Context, path string, endOffset int64, parser *activity.LogParser, warmer SessionCorrelatorWarmer, logger Logger) error {
	if endOffset <= 0 || warmer == nil {
		return nil
	}
	if logger == nil {
		logger = Nop
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	br := bufio.NewReader(f)
	pos := int64(0)
	for pos < endOffset {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		lineBytes, err := br.ReadBytes('\n')
		if len(lineBytes) == 0 && err == io.EOF {
			break
		}
		pos += int64(len(lineBytes))
		if pos > endOffset {
			break
		}
		line := string(lineBytes)
		if err != nil && err != io.EOF {
			return err
		}
		lineTrimmed := trimNL(line)
		if lineTrimmed != "" {
			if _, parseErr := dispatchOutputLogLine(lineTrimmed, parser, EventHandlerFunc(warmer.WarmFromParsedEvent)); parseErr != nil {
				logDispatchLineErr(logger, parseErr,
					"[logwatcher] warm parse error: %v", "[logwatcher] warm dispatch error: %v")
			}
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

// ProgressCallback receives absolute byte offset in the file after each line (including newline) and the raw line text.
type ProgressCallback func(byteOffset int64, line string)

// ProcessOutputLogFileFromOffset reads from startOffset and returns the final byte offset in the file.
func ProcessOutputLogFileFromOffset(ctx context.Context, path string, startOffset int64, parser *activity.LogParser, handler EventHandler, logger Logger, onProgress ProgressCallback) (int64, error) {
	if logger == nil {
		logger = Nop
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
			if _, parseErr := dispatchOutputLogLine(lineTrimmed, parser, handler); parseErr != nil {
				logDispatchLineErr(logger, parseErr,
					"[logwatcher] bootstrap parse error: %v", "[logwatcher] bootstrap dispatch error: %v")
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
