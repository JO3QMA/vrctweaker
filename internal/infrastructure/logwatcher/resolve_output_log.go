package logwatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrNoOutputLogFiles is returned when a directory has no matching output_log*.txt files.
var ErrNoOutputLogFiles = fmt.Errorf("no output_log*.txt files in directory")

// isVRChatPrimaryOutputLogFile is false for auxiliary names such as
// output_log_YYYY-MM-DD_HH-MM-SS.parsed_lines.txt (extract-parsed-lines output), which
// still match glob output_log*.txt.
func isVRChatPrimaryOutputLogFile(path string) bool {
	return !strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".parsed_lines.txt")
}

// ListOutputLogFiles returns all output_log*.txt paths under dir sorted by name (ascending).
func ListOutputLogFiles(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "output_log*.txt"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var out []string
	for _, p := range matches {
		if !isVRChatPrimaryOutputLogFile(p) {
			continue
		}
		info, statErr := os.Stat(p)
		if statErr != nil || !info.Mode().IsRegular() {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, ErrNoOutputLogFiles
	}
	return out, nil
}

// OutputLogPathValid reports whether path is an existing directory suitable for log watching.
// Empty directories are valid (logs may appear later). Regular files are rejected (ADR 0005 Decision 14).
func OutputLogPathValid(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
