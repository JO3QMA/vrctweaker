package appversion

import (
	"encoding/json"
	"fmt"
	"strings"
)

const issuesURL = "https://github.com/JO3QMA/vrctweaker/issues"

// Version is the application release version (from wails.json at startup, or "dev" in tests).
var Version = "dev"

// SetVersion overrides Version (tests only).
func SetVersion(v string) {
	v = strings.TrimSpace(v)
	if v != "" {
		Version = v
	}
}

// UserAgent returns the VRChat API User-Agent string.
func UserAgent() string {
	return fmt.Sprintf("VRChat Tweaker/%s (+%s)", Version, issuesURL)
}

// ParseProductVersionFromWailsJSON reads info.productVersion from wails.json bytes.
func ParseProductVersionFromWailsJSON(data []byte) (string, error) {
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return strings.TrimSpace(cfg.Info.ProductVersion), nil
}

// ApplyWailsJSON sets Version from embedded wails.json when productVersion is present.
// On JSON parse error it returns the error and leaves Version unchanged (default "dev").
func ApplyWailsJSON(data []byte) error {
	ver, err := ParseProductVersionFromWailsJSON(data)
	if err != nil {
		return err
	}
	if ver != "" {
		Version = ver
	}
	return nil
}
