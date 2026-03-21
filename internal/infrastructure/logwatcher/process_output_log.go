package logwatcher

import (
	"bufio"
	"context"
	"os"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// ProcessOutputLogFile reads an entire output_log file from the beginning and dispatches parsed events.
// Used once to bootstrap when the activity tables are empty.
func ProcessOutputLogFile(ctx context.Context, path string, parser LogParser, handler EventHandler, logger Logger) error {
	if logger == nil {
		logger = nopLogger{}
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)

	for sc.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := trimNL(sc.Text())
		if line == "" {
			continue
		}
		baseTime := activity.ParseVRChatTimestamp(line, time.Now().In(time.Local))
		events, parseErr := parser.ParseLine(line, baseTime)
		if parseErr != nil {
			logger.Printf("[logwatcher] bootstrap parse error: %v", parseErr)
			continue
		}
		for _, ev := range events {
			if ev != nil {
				handler.Handle(ev)
			}
		}
	}
	return sc.Err()
}
