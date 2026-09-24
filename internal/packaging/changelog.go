package packaging

import (
	"fmt"
	"strings"
)

// ReleaseNotesFromChangelog extracts the section for version (e.g. "0.1.0") from CHANGELOG.md bytes.
func ReleaseNotesFromChangelog(data []byte, version string) (string, error) {
	ver := strings.TrimSpace(version)
	if ver == "" {
		return "", fmt.Errorf("packaging: changelog version is required")
	}
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	var body []string
	capturing := false
	for _, line := range lines {
		if strings.HasPrefix(line, "## [") {
			closeIdx := strings.Index(line, "]")
			if closeIdx < 0 {
				continue
			}
			sectionVer := strings.TrimSpace(line[4:closeIdx])
			if capturing {
				break
			}
			if sectionVer == ver {
				capturing = true
			}
			continue
		}
		if capturing {
			body = append(body, line)
		}
	}
	if !capturing {
		return "", fmt.Errorf("packaging: no changelog section for version %q", ver)
	}
	text := strings.TrimSpace(strings.Join(body, "\n"))
	if text == "" {
		return fmt.Sprintf("## %s\n\n_No changelog entry._\n", ver), nil
	}
	return fmt.Sprintf("## %s\n\n%s\n", ver, text), nil
}
