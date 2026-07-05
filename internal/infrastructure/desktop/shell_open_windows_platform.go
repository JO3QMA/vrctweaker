//go:build windows

package desktop

import (
	"fmt"
	"os/exec"
)

func openFolderPlatform(abs string) error {
	cmd := exec.Command("explorer", abs)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open folder: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func openFilePlatform(abs string) error {
	return openFileWindows(abs)
}

func revealPlatform(abs string) error {
	return revealWindowsExec(abs)
}
