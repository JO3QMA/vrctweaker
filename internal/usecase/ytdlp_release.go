package usecase

import (
	"context"
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

	"vrchat-tweaker/internal/appversion"
	"vrchat-tweaker/internal/platform/errtext"
)

// DefaultYTDLPReleasesLatestURL is the GitHub API URL for the latest yt-dlp release.
const DefaultYTDLPReleasesLatestURL = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"

const ytdlpExeName = "yt-dlp.exe"

const officialYTDlpRepoReleaseDownloadPrefix = "/yt-dlp/yt-dlp/releases/download/"

// YTDLPReleaseInfo is the latest official Windows exe from GitHub Releases.
type YTDLPReleaseInfo struct {
	Tag          string
	Version      string
	DownloadURL  string
	ChecksumsURL string
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
	SkipChecksumVerification  bool // tests only
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
	tag, url, _, err := ytdlpReleaseAssetsFromReleaseJSON(data)
	return tag, url, err
}

// ytdlpReleaseTagFromDownloadURL extracts the release tag from a GitHub browser_download_url.
func ytdlpReleaseTagFromDownloadURL(raw string) string {
	const marker = "/releases/download/"
	i := strings.Index(raw, marker)
	if i < 0 {
		return ""
	}
	rest := raw[i+len(marker):]
	end := strings.Index(rest, "/")
	if end <= 0 {
		return ""
	}
	return appversion.NormalizeProductVersion(rest[:end])
}

var allowedYTDlpDownloadHosts = map[string]struct{}{
	"github.com":                           {},
	"objects.githubusercontent.com":        {},
	"release-assets.githubusercontent.com": {}, // github.com/releases/download → 302
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
	if host == "github.com" {
		return validateOfficialYTDlpGitHubDownloadPath(parsed.Path)
	}
	return validateYTDlpCDNAssetPath(parsed.Path)
}

func validateOfficialYTDlpGitHubDownloadPath(path string) error {
	if !strings.HasPrefix(path, officialYTDlpRepoReleaseDownloadPrefix) {
		return errors.New("download URL is not from official yt-dlp release")
	}
	rest := strings.TrimPrefix(path, officialYTDlpRepoReleaseDownloadPrefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return errors.New("download URL is not from official yt-dlp release")
	}
	tag := strings.TrimSpace(parts[0])
	file := strings.TrimSpace(parts[1])
	if tag == "" {
		return errors.New("download URL is not from official yt-dlp release")
	}
	switch file {
	case ytdlpExeName, ytdlpSHA256SumsAssetName:
		return nil
	default:
		return errors.New("download URL is not from official yt-dlp release asset")
	}
}

func validateYTDlpCDNAssetPath(path string) error {
	base := filepath.Base(path)
	switch base {
	case ytdlpExeName, ytdlpSHA256SumsAssetName:
		return nil
	default:
		return errors.New("download URL is not an official yt-dlp release asset")
	}
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
	req.Header.Set("User-Agent", appversion.UserAgent())

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
		return YTDLPReleaseInfo{}, fmt.Errorf("github api: %s: %s", resp.Status, errtext.TruncateForErr(string(body), 200))
	}
	tag, exeURL, sumsURL, err := ytdlpReleaseAssetsFromReleaseJSON(body)
	if err != nil {
		return YTDLPReleaseInfo{}, err
	}
	return YTDLPReleaseInfo{
		Tag:          strings.TrimSpace(tag),
		Version:      appversion.NormalizeProductVersion(tag),
		DownloadURL:  exeURL,
		ChecksumsURL: sumsURL,
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

// DownloadToCache downloads an official release into Official yt-dlp cache (atomic replace via .partial).
func (u *YTDLPUpdater) DownloadToCache(ctx context.Context, cachePath string, release YTDLPReleaseInfo) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.downloadReleaseToCacheLocked(ctx, cachePath, release)
}

func ytdlpCacheExePresent(cachePath string) bool {
	st, err := os.Stat(cachePath)
	return err == nil && !st.IsDir()
}

// EnsureOfficialCache downloads the latest release into cache when cache is missing.
func (u *YTDLPUpdater) EnsureOfficialCache(ctx context.Context, cachePath string) (YTDLPReleaseInfo, error) {
	u.mu.Lock()
	if ytdlpCacheExePresent(cachePath) {
		ver := LocalYTDLPVersion(ctx, cachePath)
		u.mu.Unlock()
		return YTDLPReleaseInfo{Version: ver, Tag: ver}, nil
	}
	u.mu.Unlock()

	info, err := u.FetchLatestRelease(ctx)
	if err != nil {
		if ytdlpCacheExePresent(cachePath) {
			ver := LocalYTDLPVersion(ctx, cachePath)
			return YTDLPReleaseInfo{Version: ver, Tag: ver}, nil
		}
		return YTDLPReleaseInfo{}, err
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	if ytdlpCacheExePresent(cachePath) {
		ver := LocalYTDLPVersion(ctx, cachePath)
		if ver == "" {
			ver = info.Version
		}
		return YTDLPReleaseInfo{
			Tag:         info.Tag,
			Version:     ver,
			DownloadURL: info.DownloadURL,
		}, nil
	}
	if err := u.downloadReleaseToCacheLocked(ctx, cachePath, info); err != nil {
		return YTDLPReleaseInfo{}, err
	}
	if info.Version == "" {
		info.Version = LocalYTDLPVersion(ctx, cachePath)
	}
	return info, nil
}

func (u *YTDLPUpdater) downloadHTTPClient() *http.Client {
	base := u.httpClient()
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		return validateYTDlpDownloadURL(u, req.URL.String())
	}
	if base.CheckRedirect != nil {
		orig := base.CheckRedirect
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			if err := orig(req, via); err != nil {
				return err
			}
			return validateYTDlpDownloadURL(u, req.URL.String())
		}
	}
	c := *base
	c.CheckRedirect = checkRedirect
	return &c
}

func (u *YTDLPUpdater) downloadReleaseToCacheLocked(ctx context.Context, cachePath string, release YTDLPReleaseInfo) error {
	downloadURL := strings.TrimSpace(release.DownloadURL)
	if err := validateYTDlpDownloadURL(u, downloadURL); err != nil {
		return err
	}
	sumsURL := strings.TrimSpace(release.ChecksumsURL)
	if sumsURL == "" {
		if u == nil || !u.SkipChecksumVerification {
			return errors.New("release checksums URL is empty")
		}
	}
	if sumsURL != "" {
		if err := validateYTDlpDownloadURL(u, sumsURL); err != nil {
			return err
		}
	}
	if st, err := os.Stat(cachePath); err == nil && st.IsDir() {
		return fmt.Errorf("cache path is a directory: %s", cachePath)
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	part := cachePath + ".partial"
	if err := downloadToFile(ctx, u.downloadHTTPClient(), downloadURL, part); err != nil {
		_ = os.Remove(part)
		return err
	}
	if u != nil && u.SkipChecksumVerification {
		return u.installDownloadedCache(part, cachePath)
	}
	sumsBody, err := fetchURLBody(ctx, u.downloadHTTPClient(), sumsURL, 4<<20)
	if err != nil {
		_ = os.Remove(part)
		return fmt.Errorf("fetch SHA2-256SUMS: %w", err)
	}
	expected, err := parseYTDlpSHA256SumForFile(sumsBody, ytdlpExeName)
	if err != nil {
		_ = os.Remove(part)
		return err
	}
	got, err := sha256HexFile(part)
	if err != nil {
		_ = os.Remove(part)
		return err
	}
	if !strings.EqualFold(got, expected) {
		_ = os.Remove(part)
		return fmt.Errorf("yt-dlp.exe checksum mismatch")
	}
	return u.installDownloadedCache(part, cachePath)
}

func (u *YTDLPUpdater) installDownloadedCache(part, cachePath string) error {
	bak := cachePath + ".bak"
	_ = os.Remove(bak)
	if err := os.Rename(cachePath, bak); err != nil && !os.IsNotExist(err) {
		_ = os.Remove(part)
		return fmt.Errorf("backup cache: %w", err)
	}
	if err := os.Rename(part, cachePath); err != nil {
		_ = os.Rename(bak, cachePath)
		_ = os.Remove(part)
		return fmt.Errorf("install cache: %w", err)
	}
	_ = os.Remove(bak)
	return nil
}

func fetchURLBody(ctx context.Context, client *http.Client, rawURL string, maxBytes int64) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", appversion.UserAgent())
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("download failed (%s): %s", resp.Status, errtext.TruncateForErr(string(slurp), 200))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return body, nil
}

const maxYTDlpDownloadBytes = 100 << 20

func downloadToFile(ctx context.Context, client *http.Client, url, dest string) error {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", appversion.UserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("download failed (%s): %s", resp.Status, errtext.TruncateForErr(string(slurp), 200))
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	n, err := io.Copy(f, io.LimitReader(resp.Body, maxYTDlpDownloadBytes+1))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(dest)
		return fmt.Errorf("write failed: %w", err)
	}
	if n > maxYTDlpDownloadBytes {
		_ = os.Remove(dest)
		return fmt.Errorf("download exceeds %d bytes", maxYTDlpDownloadBytes)
	}
	return nil
}
