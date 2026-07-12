package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultYTDLPReleasesLatestURL is the GitHub API URL for the latest yt-dlp release.
const DefaultYTDLPReleasesLatestURL = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"

const userAgentGitHub = "VRChatTweaker/1.0.0 (+https://github.com/JO3QMA/vrctweaker)"

const ytdlpExeName = "yt-dlp.exe"

// YTDLPReleaseInfo is the latest official Windows exe from GitHub Releases.
type YTDLPReleaseInfo struct {
	Tag         string
	Version     string
	DownloadURL string
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubReleaseLatest struct {
	TagName string               `json:"tag_name"`
	Assets  []githubReleaseAsset `json:"assets"`
}

// YTDLPUpdater fetches official yt-dlp releases into the Official yt-dlp cache.
type YTDLPUpdater struct {
	mu                        sync.Mutex
	HTTPClient                *http.Client
	ReleasesLatestURL         string
	SkipDownloadURLValidation bool // tests only
}

// NewYTDLPUpdater returns an updater with sane defaults.
func NewYTDLPUpdater() *YTDLPUpdater {
	return &YTDLPUpdater{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		ReleasesLatestURL: DefaultYTDLPReleasesLatestURL,
	}
}

func (u *YTDLPUpdater) httpClient() *http.Client {
	if u.HTTPClient != nil {
		return u.HTTPClient
	}
	return http.DefaultClient
}

func ytdlpExeAssetFromReleaseJSON(data []byte) (tagName, downloadURL string, err error) {
	var rel githubReleaseLatest
	if err := json.Unmarshal(data, &rel); err != nil {
		return "", "", fmt.Errorf("parse release JSON: %w", err)
	}
	tag := strings.TrimSpace(rel.TagName)
	if tag == "" {
		return "", "", errors.New("release has empty tag_name")
	}
	for _, a := range rel.Assets {
		if a.Name == ytdlpExeName && strings.TrimSpace(a.BrowserDownloadURL) != "" {
			return tag, a.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("asset %q not found in release", ytdlpExeName)
}

func normalizeReleaseTag(tag string) string {
	return strings.TrimPrefix(strings.TrimSpace(tag), "v")
}

var allowedYTDlpDownloadHosts = map[string]struct{}{
	"github.com":                    {},
	"objects.githubusercontent.com": {},
}

func validateYTDlpDownloadURL(u *YTDLPUpdater, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("download URL is empty")
	}
	if u != nil && u.SkipDownloadURLValidation {
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
	if _, ok := allowedYTDlpDownloadHosts[host]; !ok {
		return fmt.Errorf("download URL host not allowed: %s", host)
	}
	return nil
}

// FetchLatestRelease resolves the latest release tag and yt-dlp.exe browser_download_url.
func (u *YTDLPUpdater) FetchLatestRelease(ctx context.Context) (YTDLPReleaseInfo, error) {
	apiURL := u.ReleasesLatestURL
	if strings.TrimSpace(apiURL) == "" {
		apiURL = DefaultYTDLPReleasesLatestURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return YTDLPReleaseInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgentGitHub)

	resp, err := u.httpClient().Do(req)
	if err != nil {
		return YTDLPReleaseInfo{}, fmt.Errorf("github api request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return YTDLPReleaseInfo{}, fmt.Errorf("read release response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return YTDLPReleaseInfo{}, fmt.Errorf("github api: %s: %s", resp.Status, truncateForErr(string(body), 200))
	}
	tag, url, err := ytdlpExeAssetFromReleaseJSON(body)
	if err != nil {
		return YTDLPReleaseInfo{}, err
	}
	return YTDLPReleaseInfo{
		Tag:         strings.TrimSpace(tag),
		Version:     normalizeReleaseTag(tag),
		DownloadURL: url,
	}, nil
}

// LocalYTDLPVersion returns a display version for yt-dlp.exe at exePath (Windows VERSIONINFO).
func LocalYTDLPVersion(ctx context.Context, exePath string) string {
	_ = ctx
	if exePath == "" {
		return ""
	}
	st, err := os.Stat(exePath)
	if err != nil || st.IsDir() {
		return ""
	}
	return localYTDLPFileVersionString(exePath)
}

// DownloadToCache downloads downloadURL into Official yt-dlp cache (atomic replace via .partial).
func (u *YTDLPUpdater) DownloadToCache(ctx context.Context, cachePath, downloadURL string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.downloadToCacheLocked(ctx, cachePath, downloadURL)
}

// EnsureOfficialCache downloads the latest release into cache when cache is missing.
func (u *YTDLPUpdater) EnsureOfficialCache(ctx context.Context, cachePath string) (YTDLPReleaseInfo, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if st, err := os.Stat(cachePath); err == nil && !st.IsDir() {
		return YTDLPReleaseInfo{Version: LocalYTDLPVersion(ctx, cachePath)}, nil
	}
	info, err := u.FetchLatestRelease(ctx)
	if err != nil {
		return YTDLPReleaseInfo{}, err
	}
	if err := u.downloadToCacheLocked(ctx, cachePath, info.DownloadURL); err != nil {
		return YTDLPReleaseInfo{}, err
	}
	if info.Version == "" {
		info.Version = LocalYTDLPVersion(ctx, cachePath)
	}
	return info, nil
}

func (u *YTDLPUpdater) downloadToCacheLocked(ctx context.Context, cachePath, downloadURL string) error {
	if err := validateYTDlpDownloadURL(u, downloadURL); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	part := cachePath + ".partial"
	if err := downloadToFile(ctx, u.httpClient(), downloadURL, part); err != nil {
		_ = os.Remove(part)
		return err
	}
	_ = os.Remove(cachePath)
	if err := os.Rename(part, cachePath); err != nil {
		_ = os.Remove(part)
		return fmt.Errorf("install cache: %w", err)
	}
	return nil
}

func downloadToFile(ctx context.Context, client *http.Client, url, dest string) error {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgentGitHub)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("download failed (%s): %s", resp.Status, truncateForErr(string(slurp), 200))
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return fmt.Errorf("write failed: %w", err)
	}
	return f.Close()
}

func truncateForErr(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
