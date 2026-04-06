//go:build windows

package sleepsuppress

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

var (
	kernel32                    = windows.NewLazySystemDLL("kernel32.dll")
	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

type windowsExecutionState struct{}

// NewExecutionState applies SetThreadExecutionState on Windows.
func NewExecutionState() ExecutionState {
	return windowsExecutionState{}
}

func (windowsExecutionState) SetSuppress(on bool) error {
	if err := kernel32.Load(); err != nil {
		return err
	}
	var flags uintptr
	if on {
		flags = esContinuous | esSystemRequired | esDisplayRequired
	} else {
		flags = esContinuous
	}
	r, _, e := procSetThreadExecutionState.Call(flags)
	if r == 0 {
		if e != nil && e != syscall.Errno(0) {
			return fmt.Errorf("SetThreadExecutionState: %w", e)
		}
		return fmt.Errorf("SetThreadExecutionState failed")
	}
	return nil
}
