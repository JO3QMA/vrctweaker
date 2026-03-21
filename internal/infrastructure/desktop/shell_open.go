package desktop

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OpenFileWithDefaultApp opens an existing regular file with the OS default application.
func OpenFileWithDefaultApp(path string) error {
	abs, err := ValidateRegularFile(path)
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		return openFileWindows(abs)
	case "darwin":
		if err := exec.Command("open", abs).Run(); err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		return nil
	default:
		return openFileXDG(abs)
	}
}

// RevealInFileManager opens a file manager focused on the given file when the OS supports it.
func RevealInFileManager(path string) error {
	abs, err := ValidateRegularFile(path)
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		return revealWindowsExec(abs)
	case "darwin":
		if err := exec.Command("open", "-R", abs).Run(); err != nil {
			return fmt.Errorf("reveal in file manager: %w", err)
		}
		return nil
	default:
		return revealLinux(abs)
	}
}

func openFileWindows(abs string) error {
	// cmd /c start "" <path> — empty title for paths with spaces
	cmd := exec.Command("cmd", "/c", "start", "", abs)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	return nil
}

func revealWindows(abs string) string {
	// explorer /select,<path> — quote when path contains spaces
	arg := "/select," + abs
	if strings.ContainsAny(abs, " \t") {
		arg = `/select,"` + abs + `"`
	}
	return arg
}

func revealWindowsExec(abs string) error {
	arg := revealWindows(abs)
	cmd := exec.Command("explorer", arg)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("reveal in file manager: %w", err)
	}
	// explorer.exe often exits with status 1 even when the window opened successfully;
	// do not Run()/Wait() in the caller or the UI would show a spurious error.
	go func() { _ = cmd.Wait() }()
	return nil
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
