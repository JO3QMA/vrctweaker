package appversion

import "strings"

// NormalizeProductVersion trims space and a leading "v" for comparisons and storage.
func NormalizeProductVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
