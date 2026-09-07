package logwatcher

import (
	"bufio"
	"context"
	"io"
)

// ScannedLine is one newline-delimited line read from an output_log file.
type ScannedLine struct {
	Raw        string
	Trimmed    string
	ByteOffset int64 // absolute file offset after consuming this line (including newline)
}

// scanOutputLogFromReader reads lines from br, which is already positioned at startOffset.
// When endOffset > 0, scanning stops once the byte position exceeds endOffset.
// onLine is called for each line with a non-empty Trimmed value.
// Returns the final byte offset in the file.
func scanOutputLogFromReader(
	ctx context.Context,
	br *bufio.Reader,
	startOffset int64,
	endOffset int64,
	onLine func(ScannedLine) error,
) (int64, error) {
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
		if endOffset > 0 && pos > endOffset {
			return endOffset, nil
		}

		line := string(lineBytes)
		if err != nil && err != io.EOF {
			return pos, err
		}

		trimmed := trimNL(line)
		if trimmed != "" && onLine != nil {
			if cbErr := onLine(ScannedLine{
				Raw:        line,
				Trimmed:    trimmed,
				ByteOffset: pos,
			}); cbErr != nil {
				return pos, cbErr
			}
		}

		if err == io.EOF {
			break
		}
	}
	return pos, nil
}
