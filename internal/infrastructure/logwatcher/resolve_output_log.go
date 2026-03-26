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

type outputLogCandidate struct {
	path    string
	modUnix int64
}

// ResolveLatestOutputLogFile returns the path to the newest output_log*.txt under dir (by ModTime, then name).
func ResolveLatestOutputLogFile(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "output_log*.txt"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", ErrNoOutputLogFiles
	}
	var cands []outputLogCandidate
	for _, p := range matches {
		if !isVRChatPrimaryOutputLogFile(p) {
			continue
		}
		info, statErr := os.Stat(p)
		if statErr != nil || !info.Mode().IsRegular() {
			continue
		}
		cands = append(cands, outputLogCandidate{path: p, modUnix: info.ModTime().UnixNano()})
	}
	if len(cands) == 0 {
		return "", ErrNoOutputLogFiles
	}
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].modUnix != cands[j].modUnix {
			return cands[i].modUnix < cands[j].modUnix
		}
		return cands[i].path < cands[j].path
	})
	return cands[len(cands)-1].path, nil
}

// OutputLogPathValid reports whether path is an existing regular file or a directory
// that contains at least one output_log*.txt file.
func OutputLogPathValid(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info == nil {
		return false
	}
	if info.Mode().IsRegular() {
		return true
	}
	if info.IsDir() {
		_, err := ResolveLatestOutputLogFile(path)
		return err == nil
	}
	return false
}
