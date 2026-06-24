package logwatcher

import (
	"context"
	"time"
)

// VRChatRunningChecker reports whether the VRChat client process is running.
type VRChatRunningChecker interface {
	VRChatRunning() (bool, error)
}

// MonitorVRChatRunning polls until ctx is cancelled. onStopped is called once when VRChat
// transitions from running to not running.
func MonitorVRChatRunning(ctx context.Context, interval time.Duration, checker VRChatRunningChecker, onStopped func()) error {
	if checker == nil || onStopped == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	if interval <= 0 {
		interval = 4 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	wasRunning := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			running, err := checker.VRChatRunning()
			if err != nil {
				continue
			}
			if wasRunning && !running {
				onStopped()
			}
			wasRunning = running
		}
	}
}
