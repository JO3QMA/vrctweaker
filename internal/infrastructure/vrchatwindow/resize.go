// Package vrchatwindow resizes the running VRChat client window (Windows).
package vrchatwindow

import "errors"

var (
	// ErrUnsupported is returned off Windows.
	ErrUnsupported = errors.New("set_vrchat_window_size: unsupported platform")
	// ErrNotRunning means VRChat.exe is not running.
	ErrNotRunning = errors.New("set_vrchat_window_size: vrchat not running")
	// ErrNoWindow means a suitable top-level window was not found.
	ErrNoWindow = errors.New("set_vrchat_window_size: window not found")
	// ErrInvalidSize means width/height are not positive.
	ErrInvalidSize = errors.New("set_vrchat_window_size: width and height must be positive")
	// ErrResizeFailed means SetWindowPos did not apply the requested size.
	ErrResizeFailed = errors.New("set_vrchat_window_size: resize did not apply")
)

// Resize sets the main VRChat window to width×height (pixels), keeping position.
// Exclusive fullscreen is a no-op success (skip). Maximized windows are restored first.
func Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return ErrInvalidSize
	}
	return resize(width, height)
}
