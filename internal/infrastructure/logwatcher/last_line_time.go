package logwatcher

import (
	"bufio"
	"os"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// LastVRChatLineTimeInFile returns the timestamp from the last line in path that starts with
// a VRChat output_log timestamp, or zero time if none.
func LastVRChatLineTimeInFile(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = f.Close() }()

	var last time.Time
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if ts := activity.ParseVRChatTimestamp(sc.Text(), time.Time{}); !ts.IsZero() {
			last = ts
		}
	}
	if err := sc.Err(); err != nil {
		return time.Time{}, err
	}
	return last, nil
}
