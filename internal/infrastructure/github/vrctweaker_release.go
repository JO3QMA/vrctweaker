package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"vrchat-tweaker/internal/appversion"
)

const DefaultVRCTweakerReleasesLatestURL = "https://api.github.com/repos/JO3QMA/vrctweaker/releases/latest"

const windowsZipSuffix = "-windows-amd64.zip"

var zipSHA256FromBody = regexp.MustCompile(`(?i)SHA256\s*\(zip\):\s*(?:\*\*)?\s*` + "`" + `([0-9a-f]{64})` + "`")

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releaseLatest struct {
	TagName    string         `json:"tag_name"`
	Body       string         `json:"body"`
	HTMLURL    string         `json:"html_url"`
	Prerelease bool           `json:"prerelease"`
	Draft      bool           `json:"draft"`
	Assets     []releaseAsset `json:"assets"`
}

// VRCTweakerRelease is a parsed stable GitHub release for the Windows zip updater.
type VRCTweakerRelease struct {
	Tag         string
	Version     string
	Body        string
	HTMLURL     string
	DownloadURL string
	ZipSHA256   string
}

// ReleaseClient fetches VRChat Tweaker releases from GitHub.
type ReleaseClient struct {
	HTTPClient        *http.Client
	ReleasesLatestURL string
	UserAgent         string
	SkipURLValidation bool // tests only
}

func (c *ReleaseClient) latestURL() string {
	if c != nil && strings.TrimSpace(c.ReleasesLatestURL) != "" {
		return strings.TrimSpace(c.ReleasesLatestURL)
	}
	return DefaultVRCTweakerReleasesLatestURL
}

func (c *ReleaseClient) ua() string {
	if c != nil && strings.TrimSpace(c.UserAgent) != "" {
		return strings.TrimSpace(c.UserAgent)
	}
	return appversion.UserAgent()
}

const maxReleaseAPIResponseBytes = 1 << 20 // 1 MiB

func (c *ReleaseClient) defaultHTTPClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// FetchLatestStable returns the latest non-prerelease release with Windows zip asset and body SHA256.
func (c *ReleaseClient) FetchLatestStable(ctx context.Context) (VRCTweakerRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.latestURL(), nil)
	if err != nil {
		return VRCTweakerRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.ua())

	resp, err := c.defaultHTTPClient().Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return VRCTweakerRelease{}, fmt.Errorf("github releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return VRCTweakerRelease{}, fmt.Errorf("github releases: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(slurp)))
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxReleaseAPIResponseBytes))
	if err != nil {
		return VRCTweakerRelease{}, fmt.Errorf("github releases: read body: %w", err)
	}
	return parseLatestReleaseJSON(data, c != nil && c.SkipURLValidation)
}

func parseLatestReleaseJSON(data []byte, skipURLValidation bool) (VRCTweakerRelease, error) {
	var rel releaseLatest
	if err := json.Unmarshal(data, &rel); err != nil {
		return VRCTweakerRelease{}, fmt.Errorf("parse release JSON: %w", err)
	}
	if rel.Draft {
		return VRCTweakerRelease{}, errors.New("release is draft")
	}
	if rel.Prerelease {
		return VRCTweakerRelease{}, errors.New("release is prerelease")
	}
	tag := strings.TrimSpace(rel.TagName)
	if tag == "" {
		return VRCTweakerRelease{}, errors.New("release has empty tag_name")
	}
	version := strings.TrimPrefix(tag, "v")
	if err := appversion.ValidateProductVersion(version); err != nil {
		return VRCTweakerRelease{}, fmt.Errorf("invalid release version: %w", err)
	}
	zipURL, err := windowsZipAssetURL(rel.Assets)
	if err != nil {
		return VRCTweakerRelease{}, err
	}
	if valErr := validateDownloadURL(zipURL, skipURLValidation); valErr != nil {
		return VRCTweakerRelease{}, valErr
	}
	sha, err := parseZipSHA256FromBody(rel.Body)
	if err != nil {
		return VRCTweakerRelease{}, err
	}
	html := strings.TrimSpace(rel.HTMLURL)
	if html == "" {
		html = "https://github.com/JO3QMA/vrctweaker/releases/latest"
	}
	return VRCTweakerRelease{
		Tag:         tag,
		Version:     version,
		Body:        rel.Body,
		HTMLURL:     html,
		DownloadURL: zipURL,
		ZipSHA256:   sha,
	}, nil
}

func windowsZipAssetURL(assets []releaseAsset) (string, error) {
	for _, a := range assets {
		name := strings.TrimSpace(a.Name)
		if name != "" && strings.HasSuffix(strings.ToLower(name), windowsZipSuffix) {
			url := strings.TrimSpace(a.BrowserDownloadURL)
			if url != "" {
				return url, nil
			}
		}
	}
	return "", fmt.Errorf("windows zip asset %q not found", windowsZipSuffix)
}

func parseZipSHA256FromBody(body string) (string, error) {
	m := zipSHA256FromBody.FindStringSubmatch(body)
	if len(m) < 2 {
		return "", errors.New("zip SHA256 not found in release body")
	}
	return strings.ToLower(m[1]), nil
}

// TruncateReleaseNotes returns plain-text notes capped at maxRunes characters.
func TruncateReleaseNotes(body string, maxRunes int) string {
	body = strings.TrimSpace(body)
	if maxRunes <= 0 || body == "" {
		return ""
	}
	runes := []rune(body)
	if len(runes) <= maxRunes {
		return body
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "…"
}
