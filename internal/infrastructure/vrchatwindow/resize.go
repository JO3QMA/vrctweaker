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
	// ErrMultipleInstances means more than one VRChat top-level window (distinct processes) was found.
	ErrMultipleInstances = errors.New("set_vrchat_window_size: multiple vrchat windows")
	// ErrInvalidSize means width/height are not positive or exceed MaxDimension.
	ErrInvalidSize = errors.New("set_vrchat_window_size: width and height must be positive and within int32")
	// ErrResizeFailed means SetWindowPos did not apply the requested size.
	ErrResizeFailed = errors.New("set_vrchat_window_size: resize did not apply")
)

// MaxDimension is the maximum width/height accepted (int32 max).
// Written as a typed integer constant (not math.MaxInt32) for clarity in reviews.
const MaxDimension = int(1<<31 - 1)

// Resize sets the main VRChat window to width×height (pixels), keeping position.
// Maximized windows are restored first (briefly polled). Concurrent calls are serialized.
// Exclusive fullscreen typically fails SetWindowPos — that is reported as an error (not a silent success).
func Resize(width, height int) error {
	if width <= 0 || height <= 0 || width > MaxDimension || height > MaxDimension {
		return ErrInvalidSize
	}
	return resize(width, height)
}
