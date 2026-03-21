//go:build windows

package desktop

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func openFileWindows(abs string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	file, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	if err := windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL); err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	return nil
}
