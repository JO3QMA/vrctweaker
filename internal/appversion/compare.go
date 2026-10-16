package appversion

import (
	"strconv"
	"strings"
)

type versionCore struct {
	parts      [3]int
	prerelease bool
	invalid    bool
}

// IsDevelopmentVersion reports dev/empty builds that should not prompt for self-update.
func IsDevelopmentVersion(version string) bool {
	v := strings.TrimSpace(version)
	return v == "" || v == "dev"
}

// CompareProductVersion compares two product version strings (e.g. 1.2.3).
// Returns -1 if a < b, 0 if equal, 1 if a > b.
// Non-semver or "dev" current is treated as older than any valid release version.
// A prerelease (suffix after -/+) is older than the same numeric version without suffix.
func CompareProductVersion(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == b {
		return 0
	}
	if a == "" || a == "dev" {
		return -1
	}
	if b == "" || b == "dev" {
		return 1
	}
	ac := parseVersionCore(a)
	bc := parseVersionCore(b)
	if ac.invalid && bc.invalid {
		return 0
	}
	if ac.invalid {
		return -1
	}
	if bc.invalid {
		return 1
	}
	for i := 0; i < 3; i++ {
		if ac.parts[i] < bc.parts[i] {
			return -1
		}
		if ac.parts[i] > bc.parts[i] {
			return 1
		}
	}
	if ac.prerelease && !bc.prerelease {
		return -1
	}
	if !ac.prerelease && bc.prerelease {
		return 1
	}
	return 0
}

func parseVersionCore(v string) versionCore {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if plus := strings.Index(v, "+"); plus >= 0 {
		// Semver build metadata does not affect precedence.
		v = v[:plus]
	}
	prerelease := false
	if idx := strings.Index(v, "-"); idx >= 0 {
		prerelease = true
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	var out [3]int
	invalid := false
	for i := 0; i < len(parts); i++ {
		if i >= 3 {
			invalid = true
			break
		}
		seg := strings.TrimSpace(parts[i])
		if seg == "" {
			invalid = true
			continue
		}
		n, err := strconv.Atoi(seg)
		if err != nil || n < 0 {
			invalid = true
			continue
		}
		out[i] = n
	}
	return versionCore{parts: out, prerelease: prerelease, invalid: invalid}
}
