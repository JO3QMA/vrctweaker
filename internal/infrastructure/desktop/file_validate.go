package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateRegularFile returns the absolute path for an existing regular file.
func ValidateRegularFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	path = filepath.Clean(path)
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file")
	}
	return abs, nil
}
