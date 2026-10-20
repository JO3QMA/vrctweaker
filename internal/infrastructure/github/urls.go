package github

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

const releaseRepoDownloadPathPrefix = "/JO3QMA/vrctweaker/releases/download/"

const releaseRepoPagePathPrefix = "/JO3QMA/vrctweaker/releases"

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
	if portErr := rejectURLPort(parsed); portErr != nil {
		return portErr
	}
	if parsed.Scheme != "https" {
		return errors.New("release page URL must use https")
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" {
		return fmt.Errorf("release page host %q is not allowed", host)
	}
	cleanPath, err := cleanGitHubPath(parsed.Path)
	if err != nil {
		return err
	}
	if !pathHasPrefixSegment(cleanPath, releaseRepoPagePathPrefix) {
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
	if err := rejectURLPort(parsed); err != nil {
		return err
	}
	if parsed.Scheme != "https" {
		return errors.New("download URL must use https")
	}
	host := strings.ToLower(parsed.Hostname())
	if _, ok := allowedReleaseDownloadHosts[host]; !ok {
		return fmt.Errorf("download host %q is not allowed", host)
	}
	if host == "github.com" {
		cleanPath, err := cleanGitHubPath(parsed.Path)
		if err != nil {
			return err
		}
		if !pathHasPrefixSegment(cleanPath, releaseRepoDownloadPathPrefix) {
			return errors.New("download path is not a vrctweaker release asset")
		}
	}
	return nil
}

func rejectURLPort(u *url.URL) error {
	if u == nil {
		return errors.New("URL is nil")
	}
	if u.Port() != "" {
		return errors.New("URL must not include an explicit port")
	}
	return nil
}

func cleanGitHubPath(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", errors.New("URL path is empty")
	}
	clean := path.Clean(p)
	for _, seg := range strings.Split(strings.Trim(clean, "/"), "/") {
		if seg == ".." {
			return "", errors.New("URL path must not contain parent segments")
		}
		if seg == "." {
			return "", errors.New("URL path must not contain dot segments")
		}
	}
	return clean, nil
}

// pathHasPrefixSegment requires prefix match at a path segment boundary (not releasesX vs releases).
func pathHasPrefixSegment(cleanPath, prefix string) bool {
	if !strings.HasPrefix(cleanPath, prefix) {
		return false
	}
	if len(cleanPath) == len(prefix) {
		return true
	}
	if strings.HasSuffix(prefix, "/") {
		return true
	}
	return cleanPath[len(prefix)] == '/'
}
