package tray

import "testing"

// TestTrayWindowsRegressionNotes documents the tray freeze root cause for non-Windows CI.
// HWND creation and the GetMessage loop must share one OS thread (runtime.LockOSThread).
// The previous implementation called systray.New on the Wails goroutine and systray.Run on
// another goroutine, so the scheduler could migrate them apart and tray input stopped working.
// Explorer restarts require handling the TaskbarCreated broadcast (see w32systray).
func TestTrayWindowsRegressionNotes(t *testing.T) {
	t.Log("see internal/infrastructure/tray/tray_windows.go and w32systray/systray_windows.go")
}
