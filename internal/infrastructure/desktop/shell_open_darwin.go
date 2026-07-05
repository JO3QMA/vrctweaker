//go:build darwin

package desktop

import (
	"fmt"
	"os/exec"
)

func openFolderPlatform(abs string) error {
	if err := exec.Command("open", abs).Run(); err != nil {
		return fmt.Errorf("open folder: %w", err)
	}
	return nil
}

func openFilePlatform(abs string) error {
	if err := exec.Command("open", abs).Run(); err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	return nil
}

func revealPlatform(abs string) error {
	if err := exec.Command("open", "-R", abs).Run(); err != nil {
		return fmt.Errorf("reveal in file manager: %w", err)
	}
	return nil
}
