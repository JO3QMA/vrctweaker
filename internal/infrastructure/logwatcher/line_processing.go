package logwatcher

import (
	"errors"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

var errNilLineProcessorArg = errors.New("logwatcher: nil parser or handler")

// ParseOutputLogLine parses a trimmed non-empty output_log line into events.
// baseTime is taken from the line timestamp when present; otherwise now is used.
// This is a pure function with no handler side effects.
func ParseOutputLogLine(lineTrimmed string, parser *activity.LogParser, now time.Time) (events []activity.ParsedEvent, baseTime time.Time, err error) {
	if parser == nil {
		return nil, time.Time{}, errNilLineProcessorArg
	}
	baseTime = activity.ParseVRChatTimestamp(lineTrimmed, now.In(time.Local))
	events, err = parser.ParseLine(lineTrimmed, baseTime)
	return events, baseTime, err
}

// LineProcessor parses output_log lines and dispatches events to a handler.
type LineProcessor struct {
	Parser  *activity.LogParser
	Handler EventHandler
}

// NewLineProcessor returns a processor that dispatches parsed events to handler.
func NewLineProcessor(parser *activity.LogParser, handler EventHandler) *LineProcessor {
	return &LineProcessor{Parser: parser, Handler: handler}
}

// Process parses a trimmed non-empty line and dispatches events.
// Returns the line timestamp used for parsing (for checkpoints).
func (p *LineProcessor) Process(lineTrimmed string) (baseTime time.Time, err error) {
	if p == nil || p.Parser == nil || p.Handler == nil {
		return time.Time{}, errNilLineProcessorArg
	}
	events, baseTime, err := ParseOutputLogLine(lineTrimmed, p.Parser, time.Now())
	if err != nil {
		return baseTime, err
	}
	for _, ev := range events {
		if ev != nil {
			p.Handler.Handle(ev)
		}
	}
	return baseTime, nil
}

// logLineProcessErr logs LineProcessor errors; nil parser/handler uses nilArgFmt.
func logLineProcessErr(logger Logger, err error, parseFmt, nilArgFmt string, args ...any) {
	fmt := parseFmt
	if errors.Is(err, errNilLineProcessorArg) {
		fmt = nilArgFmt
	}
	logArgs := make([]any, len(args)+1)
	copy(logArgs, args)
	logArgs[len(args)] = err
	logger(fmt, logArgs...)
}

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
