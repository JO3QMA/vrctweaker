// Package vrchatwindow resizes the running VRChat client window (Windows).
package vrchatwindow

import (
	"errors"
	"math"
)

var (
	// ErrUnsupported is returned off Windows.
	ErrUnsupported = errors.New("set_vrchat_window_size: unsupported platform")
	// ErrNotRunning means VRChat.exe is not running.
	ErrNotRunning = errors.New("set_vrchat_window_size: vrchat not running")
	// ErrNoWindow means a suitable top-level window was not found.
	ErrNoWindow = errors.New("set_vrchat_window_size: window not found")
	// ErrMultipleInstances means more than one VRChat.exe is running.
	ErrMultipleInstances = errors.New("set_vrchat_window_size: multiple vrchat processes")
	// ErrInvalidSize means width/height are not positive or exceed int32.
	ErrInvalidSize = errors.New("set_vrchat_window_size: width and height must be positive and within int32")
	// ErrResizeFailed means SetWindowPos did not apply the requested size.
	ErrResizeFailed = errors.New("set_vrchat_window_size: resize did not apply")
)

const MaxDimension = math.MaxInt32

// Resize sets the main VRChat window to width×height (pixels), keeping position.
// Exclusive fullscreen (SetWindowPos rejected while covering the monitor) is a no-op success.
// Maximized windows are restored first. Concurrent calls are serialized.
func Resize(width, height int) error {
	if width <= 0 || height <= 0 || width > MaxDimension || height > MaxDimension {
		return ErrInvalidSize
	}
	return resize(width, height)
}
