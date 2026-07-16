package logwatcher

import (
	"errors"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

var errNilDispatchArg = errors.New("logwatcher: nil parser or handler")

// dispatchOutputLogLine parses a trimmed non-empty line and dispatches events.
// Caller logs parseErr if non-nil. baseTime is for checkpoint (ParseVRChatTimestamp result).
// parser and handler must be non-nil; otherwise errNilDispatchArg is returned.
func dispatchOutputLogLine(lineTrimmed string, parser *activity.LogParser, handler EventHandler) (baseTime time.Time, parseErr error) {
	if parser == nil || handler == nil {
		return time.Time{}, errNilDispatchArg
	}
	baseTime = activity.ParseVRChatTimestamp(lineTrimmed, time.Now().In(time.Local))
	events, parseErr := parser.ParseLine(lineTrimmed, baseTime)
	if parseErr != nil {
		return baseTime, parseErr
	}
	for _, ev := range events {
		if ev != nil {
			handler.Handle(ev)
		}
	}
	return baseTime, nil
}

// logDispatchLineErr logs dispatchOutputLogLine errors; nil parser/handler uses nilArgFmt.
func logDispatchLineErr(logger Logger, err error, parseFmt, nilArgFmt string, args ...any) {
	fmt := parseFmt
	if errors.Is(err, errNilDispatchArg) {
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
