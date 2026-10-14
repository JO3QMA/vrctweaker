//go:build !windows

package appupdate

import "fmt"

// ScheduleReplaceAndRestart is only supported on Windows.
func ScheduleReplaceAndRestart(string, string, int) error {
	return fmt.Errorf("appupdate: auto apply is Windows only")
}

// WorkDir returns a staging path (unused on non-Windows).
func WorkDir(installDir string) string {
	return installDir
}

// CleanupWorkDir is a no-op on non-Windows.
func CleanupWorkDir(string) {}

// CurrentPID returns 0 on non-Windows apply path.
func CurrentPID() int { return 0 }

// SleepBeforeQuit is a no-op on non-Windows.
func SleepBeforeQuit() {}
