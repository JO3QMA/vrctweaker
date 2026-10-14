package github

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const releaseRepoDownloadPathPrefix = "/JO3QMA/vrctweaker/releases/download/"

var allowedReleaseDownloadHosts = map[string]struct{}{
	"github.com":                           {},
	"objects.githubusercontent.com":        {},
	"release-assets.githubusercontent.com": {},
}

// ValidateReleaseDownloadURL ensures HTTPS, allowed host, and (for github.com) repo release asset path.
func ValidateReleaseDownloadURL(raw string) error {
	return validateDownloadURL(raw, false)
}

// ValidateReleasePageURL ensures HTTPS and github.com/JO3QMA/vrctweaker release page path.
func ValidateReleasePageURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("release page URL is empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid release page URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return errors.New("release page URL must use https")
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" {
		return fmt.Errorf("release page host %q is not allowed", host)
	}
	path := parsed.EscapedPath()
	if !strings.HasPrefix(path, "/JO3QMA/vrctweaker/releases") {
		return errors.New("release page path is not for this repository")
	}
	return nil
}

func validateDownloadURL(raw string, skip bool) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("download URL is empty")
	}
	if skip {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return errors.New("download URL must use https")
	}
	host := strings.ToLower(parsed.Hostname())
	if _, ok := allowedReleaseDownloadHosts[host]; !ok {
		return fmt.Errorf("download host %q is not allowed", host)
	}
	if host == "github.com" {
		if !strings.HasPrefix(parsed.EscapedPath(), releaseRepoDownloadPathPrefix) {
			return errors.New("download path is not a vrctweaker release asset")
		}
	}
	return nil
}
