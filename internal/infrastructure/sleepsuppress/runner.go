package sleepsuppress

import (
	"context"
	"runtime"
	"time"
)

// SettingGetter reads whether sleep suppression while VRChat runs is enabled.
type SettingGetter interface {
	SuppressSleepWhileVRChat(ctx context.Context) (bool, error)
}

// ProcessChecker reports whether the VRChat client process is running.
type ProcessChecker interface {
	VRChatRunning() (bool, error)
}

// ExecutionState toggles OS idle sleep / display power handling for this process (Windows).
type ExecutionState interface {
	// SetSuppress true requests continuous system (and display) availability; false clears prior flags.
	SetSuppress(on bool) error
}

// Run polls until ctx is cancelled, then clears execution state. Returns ctx.Err().
// On Windows, SetThreadExecutionState applies per OS thread; LockOSThread keeps all
// SetSuppress calls on one thread so flags are not left set after migration or shutdown.
func Run(ctx context.Context, interval time.Duration, settings SettingGetter, proc ProcessChecker, exec ExecutionState) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	suppressing := false
	clearSuppress := func() {
		if suppressing {
			_ = exec.SetSuppress(false)
			suppressing = false
		}
	}

	for {
		select {
		case <-ctx.Done():
			_ = exec.SetSuppress(false)
			return ctx.Err()
		case <-ticker.C:
			enabled, err := settings.SuppressSleepWhileVRChat(ctx)
			if err != nil {
				continue
			}
			if !enabled {
				clearSuppress()
				continue
			}
			running, err := proc.VRChatRunning()
			if err != nil {
				continue
			}
			if running {
				if err := exec.SetSuppress(true); err == nil {
					suppressing = true
				}
			} else {
				clearSuppress()
			}
		}
	}
}
