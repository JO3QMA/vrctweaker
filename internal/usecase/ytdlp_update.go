package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Name of the Windows binary in yt-dlp/yt-dlp GitHub Releases (see releases/latest assets).
const ytdlpReleaseAssetName = "yt-dlp.exe"

// DefaultYTDLPReleasesLatestURL is the GitHub API URL for the latest yt-dlp release.
const DefaultYTDLPReleasesLatestURL = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"

// userAgentGitHub is sent with requests to GitHub API and release asset downloads.
const userAgentGitHub = "VRChatTweaker/1.0.0 (+https://github.com/JO3QMA/vrctweaker)"

// VRChatYTDLPExePath returns the path VRChat uses for yt-dlp on Windows, derived from
// the same directory as config.json (…/VRChat/VRChat/config.json → …/Tools/yt-dlp.exe).
func VRChatYTDLPExePath(vrchatConfigJSONPath string) string {
	dir := filepath.Dir(vrchatConfigJSONPath)
	return filepath.Join(dir, "Tools", ytdlpReleaseAssetName)
}

// YTDLPUpdateStatus is the outcome of checking local and latest yt-dlp versions.
type YTDLPUpdateStatus struct {
	Supported         bool   `json:"supported"`
	TargetPath        string `json:"targetPath"`
	LocalVersion      string `json:"localVersion"`
	LatestVersion     string `json:"latestVersion"`
	LatestTag         string `json:"latestTag"`
	LatestDownloadURL string `json:"latestDownloadUrl"`
	LatestError       string `json:"latestError"`
	UnsupportedReason string `json:"unsupportedReason,omitempty"`
}

// YTDLPApplyResult is the outcome of downloading and installing yt-dlp.exe.
type YTDLPApplyResult struct {
	Ok             bool   `json:"ok"`
	AppliedVersion string `json:"appliedVersion"`
	Message        string `json:"message"`
	Error          string `json:"error"`
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubReleaseLatest struct {
	TagName string               `json:"tag_name"`
	Assets  []githubReleaseAsset `json:"assets"`
}

// YTDLPUpdater fetches official yt-dlp releases and installs the Windows exe.
type YTDLPUpdater struct {
	HTTPClient        *http.Client
	ReleasesLatestURL string
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

// ytdlpExeAssetFromReleaseJSON extracts tag and yt-dlp.exe download URL from a GitHub release JSON body.
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
		if a.Name == ytdlpReleaseAssetName && strings.TrimSpace(a.BrowserDownloadURL) != "" {
			return tag, a.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("asset %q not found in release", ytdlpReleaseAssetName)
}

// FetchLatestRelease resolves the latest release tag and yt-dlp.exe browser_download_url.
func (u *YTDLPUpdater) FetchLatestRelease(ctx context.Context) (tagName, downloadURL string, err error) {
	apiURL := u.ReleasesLatestURL
	if strings.TrimSpace(apiURL) == "" {
		apiURL = DefaultYTDLPReleasesLatestURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgentGitHub)

	resp, err := u.httpClient().Do(req)
	if err != nil {
		return "", "", fmt.Errorf("github api request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", "", fmt.Errorf("read release response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github api: %s: %s", resp.Status, truncateForErr(string(body), 200))
	}
	return ytdlpExeAssetFromReleaseJSON(body)
}

// LocalYTDLPVersion runs yt-dlp.exe --version when the file exists.
func LocalYTDLPVersion(ctx context.Context, exePath string) string {
	if exePath == "" {
		return ""
	}
	st, err := os.Stat(exePath)
	if err != nil || st.IsDir() {
		return ""
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, exePath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	return strings.TrimSpace(line)
}

// GetBasics returns the install path and local yt-dlp version without calling GitHub.
func (u *YTDLPUpdater) GetBasics(ctx context.Context, vrchatConfigJSONPath string) YTDLPUpdateStatus {
	target := VRChatYTDLPExePath(vrchatConfigJSONPath)
	if runtime.GOOS != "windows" {
		return YTDLPUpdateStatus{
			Supported:         false,
			TargetPath:        target,
			UnsupportedReason: "この機能は Windows 版のみ利用できます。",
		}
	}
	local := LocalYTDLPVersion(ctx, target)
	return YTDLPUpdateStatus{
		Supported:    true,
		TargetPath:   target,
		LocalVersion: local,
	}
}

// GetUpdateStatus loads local version (if any) and fetches latest from GitHub.
func (u *YTDLPUpdater) GetUpdateStatus(ctx context.Context, vrchatConfigJSONPath string) YTDLPUpdateStatus {
	st := u.GetBasics(ctx, vrchatConfigJSONPath)
	if !st.Supported {
		return st
	}
	tag, url, err := u.FetchLatestRelease(ctx)
	if err != nil {
		st.LatestError = err.Error()
		return st
	}
	st.LatestVersion = normalizeReleaseTag(tag)
	st.LatestTag = strings.TrimSpace(tag)
	st.LatestDownloadURL = url
	return st
}

func normalizeReleaseTag(tag string) string {
	return strings.TrimPrefix(strings.TrimSpace(tag), "v")
}

// ApplyLatest downloads the given URL and replaces target exe with backup.
func (u *YTDLPUpdater) ApplyLatest(ctx context.Context, vrchatConfigJSONPath, downloadURL, expectedTag string) YTDLPApplyResult {
	if runtime.GOOS != "windows" {
		return YTDLPApplyResult{
			Ok:    false,
			Error: "この機能は Windows 版のみ利用できます。",
		}
	}
	exePath := VRChatYTDLPExePath(vrchatConfigJSONPath)
	toolsDir := filepath.Dir(exePath)
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return YTDLPApplyResult{Ok: false, Error: fmt.Sprintf("Tools フォルダを作成できません: %v", err)}
	}

	partPath := exePath + ".part"
	if err := downloadToFile(ctx, u.httpClient(), downloadURL, partPath); err != nil {
		_ = os.Remove(partPath)
		return YTDLPApplyResult{Ok: false, Error: err.Error()}
	}

	if err := finishYTDLPInstall(exePath, partPath); err != nil {
		_ = os.Remove(partPath)
		return YTDLPApplyResult{
			Ok:    false,
			Error: fmt.Sprintf("既存ファイルの退避または配置に失敗しました（VRChat 終了後に再試行してください）: %v", err),
		}
	}

	applied := normalizeReleaseTag(expectedTag)
	if applied == "" {
		applied = LocalYTDLPVersion(ctx, exePath)
	}
	return YTDLPApplyResult{
		Ok:             true,
		AppliedVersion: applied,
		Message:        "適用しました。変更を反映するには VRChat を再起動してください。",
	}
}

func tryRestoreFromBak(exePath, bakPath string) error {
	if _, err := os.Stat(bakPath); err != nil {
		return err
	}
	return os.Rename(bakPath, exePath)
}

// finishYTDLPInstall moves partPath into exePath after backing up an existing exe to .bak.
func finishYTDLPInstall(exePath, partPath string) error {
	bakPath := exePath + ".bak"
	_ = os.Remove(bakPath)
	if _, err := os.Stat(exePath); err == nil {
		if err := os.Rename(exePath, bakPath); err != nil {
			return fmt.Errorf("backup: %w", err)
		}
	}
	if err := os.Rename(partPath, exePath); err != nil {
		_ = tryRestoreFromBak(exePath, bakPath)
		return fmt.Errorf("install: %w", err)
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
		return fmt.Errorf("ダウンロード要求に失敗しました: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("ダウンロードに失敗しました (%s): %s", resp.Status, truncateForErr(string(slurp), 200))
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("一時ファイルを作成できません: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return fmt.Errorf("書き込みに失敗しました: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func truncateForErr(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
