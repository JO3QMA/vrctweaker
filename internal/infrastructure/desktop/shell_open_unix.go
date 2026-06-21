//go:build !windows && !darwin

package desktop

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func openFolderPlatform(abs string) error {
	return openFileXDG(abs)
}

func openFilePlatform(abs string) error {
	return openFileXDG(abs)
}

func revealPlatform(abs string) error {
	return revealLinux(abs)
}

func openFileXDG(abs string) error {
	if _, err := exec.LookPath("xdg-open"); err != nil {
		return fmt.Errorf("xdg-open not found: %w", err)
	}
	cmd := exec.Command("xdg-open", abs)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	return nil
}

func revealLinux(abs string) error {
	dir := filepath.Dir(abs)
	if path, err := exec.LookPath("nautilus"); err == nil {
		cmd := exec.Command(path, "--select", abs)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	if _, err := exec.LookPath("xdg-open"); err == nil {
		cmd := exec.Command("xdg-open", dir)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("open folder: %w", err)
		}
		return nil
	}
	return fmt.Errorf("no supported file manager (xdg-open or nautilus)")
}
