//go:build !windows

package desktop

import "fmt"

func openFileWindows(abs string) error {
	_ = abs
	return fmt.Errorf("desktop.openFileWindows: windows-only implementation")
}
