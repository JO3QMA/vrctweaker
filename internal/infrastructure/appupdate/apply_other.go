//go:build !windows

package appupdate

import "fmt"

// ScheduleReplaceAndRestart is only supported on Windows.
// Signature matches apply_windows.go; callers live in //go:build windows usecase files only.
func ScheduleReplaceAndRestart(string, string, string, int) error {
	return fmt.Errorf("appupdate: auto apply is Windows only")
}

// WorkDir is unused on non-Windows; returns empty to avoid polluting installDir.
func WorkDir(string) string {
	return ""
}

// CleanupWorkDir is a no-op on non-Windows.
func CleanupWorkDir(string) {}

// CurrentPID returns 0 on non-Windows apply path.
func CurrentPID() int { return 0 }

// SleepBeforeQuit is a no-op on non-Windows.
func SleepBeforeQuit() {}
