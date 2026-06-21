package desktop

import (
	"fmt"
	"os/exec"
)

// OpenFolderInFileManager opens an existing directory in the OS file manager.
func OpenFolderInFileManager(dir string) error {
	abs, err := ValidateDirectory(dir)
	if err != nil {
		return err
	}
	return openFolderPlatform(abs)
}

// OpenFileWithDefaultApp opens an existing regular file with the OS default application.
func OpenFileWithDefaultApp(path string) error {
	abs, err := ValidateRegularFile(path)
	if err != nil {
		return err
	}
	return openFilePlatform(abs)
}

// RevealInFileManager opens a file manager focused on the given file when the OS supports it.
func RevealInFileManager(path string) error {
	abs, err := ValidateRegularFile(path)
	if err != nil {
		return err
	}
	return revealPlatform(abs)
}

func revealWindows(abs string) string {
	// explorer /select,<path> — Go's exec.Command handles argument quoting;
	// adding internal quotes would be double-escaped and break explorer.
	return "/select," + abs
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
