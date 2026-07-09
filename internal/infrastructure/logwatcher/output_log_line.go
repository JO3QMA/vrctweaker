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
func dispatchOutputLogLine(lineTrimmed string, parser LogParser, handler EventHandler) (baseTime time.Time, parseErr error) {
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

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
