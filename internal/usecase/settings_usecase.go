package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ErrOutputLogPathMustBeDirectory is returned when saving a non-directory output_log path.
var ErrOutputLogPathMustBeDirectory = errors.New("output log path must be a directory")

// SettingsUseCase handles app settings.
type SettingsUseCase struct {
	repo appSettingsRepo
}

// NewSettingsUseCase creates a new SettingsUseCase.
func NewSettingsUseCase(repo appSettingsRepo) *SettingsUseCase {
	return &SettingsUseCase{repo: repo}
}

// Get returns a setting value by key.
func (uc *SettingsUseCase) Get(ctx context.Context, key string) (string, error) {
	return uc.repo.Get(ctx, key)
}

// Set saves a setting.
func (uc *SettingsUseCase) Set(ctx context.Context, key, value string) error {
	return uc.repo.Set(ctx, key, value)
}

// GetLogRetentionDays returns the log retention days (default 30).
func (uc *SettingsUseCase) GetLogRetentionDays(ctx context.Context) (int, error) {
	v, err := uc.repo.Get(ctx, "log_retention_days")
	if err != nil {
		return 30, err
	}
	if v == "" {
		return 30, nil
	}
	d, err := strconv.Atoi(v)
	if err != nil || d <= 0 {
		return 30, nil
	}
	return d, nil
}

// SetLogRetentionDays saves the log retention days.
func (uc *SettingsUseCase) SetLogRetentionDays(ctx context.Context, days int) error {
	return uc.repo.Set(ctx, "log_retention_days", strconv.Itoa(days))
}

// GetOutputLogPath returns the configured VRChat log folder (empty if not set).
func (uc *SettingsUseCase) GetOutputLogPath(ctx context.Context) (string, error) {
	return uc.repo.Get(ctx, keyOutputLogPath)
}

// SaveOutputLogPath saves the VRChat log folder path (empty clears to default).
// Regular files are rejected with ErrOutputLogPathMustBeDirectory.
func (uc *SettingsUseCase) SaveOutputLogPath(ctx context.Context, path string) error {
	if err := validateOutputLogPathSetting(path); err != nil {
		return err
	}
	return uc.repo.Set(ctx, keyOutputLogPath, strings.TrimSpace(path))
}

func validateOutputLogPathSetting(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("output_log path not accessible: %w", err)
	}
	if !info.IsDir() {
		return ErrOutputLogPathMustBeDirectory
	}
	return nil
}

// EnsureOutputLogWatchDir returns the directory to watch for output_log*.txt.
// If the stored setting is a regular file, it is migrated once to the parent directory.
// Missing or unusable paths clear the setting and return ("", true, nil) so the caller can
// log a warning and fall back to the default folder. Transient I/O errors are returned
// without clearing (cleared=false). If clearing the setting fails, cleared=false and the
// Set error is returned.
func (uc *SettingsUseCase) EnsureOutputLogWatchDir(ctx context.Context) (dir string, cleared bool, err error) {
	p, err := uc.GetOutputLogPath(ctx)
	if err != nil {
		return "", false, err
	}
	p = strings.TrimSpace(p)
	if p == "" {
		return "", false, nil
	}
	absPath, absErr := filepath.Abs(filepath.Clean(p))
	if absErr != nil {
		return "", false, fmt.Errorf("output_log path cannot be resolved: %w", absErr)
	}
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return uc.clearStaleOutputLogPath(ctx)
		}
		return "", false, fmt.Errorf("output_log path not accessible: %w", statErr)
	}
	if info.IsDir() {
		return absPath, false, nil
	}
	if !info.Mode().IsRegular() {
		return uc.clearStaleOutputLogPath(ctx)
	}
	parent := filepath.Dir(absPath)
	parentInfo, parentErr := os.Stat(parent)
	if parentErr != nil {
		if errors.Is(parentErr, os.ErrNotExist) {
			return uc.clearStaleOutputLogPath(ctx)
		}
		return "", false, fmt.Errorf("output_log parent path not accessible: %w", parentErr)
	}
	if !parentInfo.IsDir() {
		return uc.clearStaleOutputLogPath(ctx)
	}
	absParent, parentAbsErr := filepath.Abs(parent)
	if parentAbsErr != nil {
		return "", false, fmt.Errorf("output_log parent path cannot be resolved: %w", parentAbsErr)
	}
	// ponytail: on Set failure keep the legacy file path and retry next startup —
	// clearing here would hide a DB write failure by silently falling back to default.
	if setErr := uc.repo.Set(ctx, keyOutputLogPath, absParent); setErr != nil {
		return "", false, setErr
	}
	return absParent, false, nil
}

func (uc *SettingsUseCase) clearStaleOutputLogPath(ctx context.Context) (string, bool, error) {
	if setErr := uc.repo.Set(ctx, keyOutputLogPath, ""); setErr != nil {
		return "", false, setErr
	}
	return "", true, nil
}

// Path settings keys in app_settings.
const (
	keyVRChatPathWindows             = "vrchat_path_windows"
	keySteamPathLinux                = "steam_path_linux"
	keyOutputLogPath                 = "output_log_path"
	keyGalleryLastExitAt             = "gallery_last_exit_at"
	keySuppressSleepWhileVRChat      = "suppress_sleep_while_vrchat"
	keyLanguage                      = "language"
	keyLastLaunchProfileID           = "last_launch_profile_id"
	keyYTDLPToolsReplaceMaintain     = "ytdlp_tools_replace_maintain"
	keyYTDLPToolsReplaceRiskAck      = "ytdlp_tools_replace_risk_ack"
	keyYTDLPToolsReplacePendingError = "ytdlp_tools_replace_pending_error"
	keyYTDLPOfficialCacheTag         = "ytdlp_official_cache_tag"
	keyYTDLPKnownLatestVersion       = "ytdlp_known_latest_version"
	keyYTDLPKnownLatestTag           = "ytdlp_known_latest_tag"
	keyYTDLPKnownLatestDownloadURL   = "ytdlp_known_latest_download_url"
)

// SupportedAppLanguages are UI locale codes persisted in app_settings.
var SupportedAppLanguages = map[string]struct{}{
	"ja":    {},
	"en":    {},
	"ko":    {},
	"zh-TW": {},
	"zh-CN": {},
}

// GetLanguage returns the saved UI language code, or empty when unset.
func (uc *SettingsUseCase) GetLanguage(ctx context.Context) (string, error) {
	return uc.repo.Get(ctx, keyLanguage)
}

// SetLanguage persists the UI language; lang must be a supported code.
func (uc *SettingsUseCase) SetLanguage(ctx context.Context, lang string) error {
	if _, ok := SupportedAppLanguages[lang]; !ok {
		return fmt.Errorf("unsupported language: %q", lang)
	}
	return uc.repo.Set(ctx, keyLanguage, lang)
}

// GetLastLaunchProfileID returns the saved Last launch profile ID, or empty when unset.
func (uc *SettingsUseCase) GetLastLaunchProfileID(ctx context.Context) (string, error) {
	return uc.repo.Get(ctx, keyLastLaunchProfileID)
}

// SetLastLaunchProfileID persists the Last launch profile ID (no existence check).
func (uc *SettingsUseCase) SetLastLaunchProfileID(ctx context.Context, profileID string) error {
	return uc.repo.Set(ctx, keyLastLaunchProfileID, strings.TrimSpace(profileID))
}

// GetGalleryLastExitAt returns the last app shutdown time used for incremental gallery sync.
// The second return is false when unset or not parseable.
func (uc *SettingsUseCase) GetGalleryLastExitAt(ctx context.Context) (time.Time, bool) {
	v, err := uc.repo.Get(ctx, keyGalleryLastExitAt)
	if err != nil || v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

// SetGalleryLastExitAt persists the shutdown instant for the next startup incremental gallery sync.
func (uc *SettingsUseCase) SetGalleryLastExitAt(ctx context.Context, t time.Time) error {
	return uc.repo.Set(ctx, keyGalleryLastExitAt, t.UTC().Format(time.RFC3339Nano))
}

// GetSuppressSleepWhileVRChat returns whether to suppress system sleep while VRChat.exe is running (Windows).
// Default is false when unset or invalid.
func (uc *SettingsUseCase) GetSuppressSleepWhileVRChat(ctx context.Context) (bool, error) {
	v, err := uc.repo.Get(ctx, keySuppressSleepWhileVRChat)
	if err != nil {
		return false, err
	}
	return parseBoolSetting(v), nil
}

// SetSuppressSleepWhileVRChat persists the sleep-suppression toggle.
func (uc *SettingsUseCase) SetSuppressSleepWhileVRChat(ctx context.Context, on bool) error {
	return uc.repo.Set(ctx, keySuppressSleepWhileVRChat, strconv.FormatBool(on))
}

// GetYTDLPToolsReplaceMaintain returns whether Tools replace maintain is desired (default false).
func (uc *SettingsUseCase) GetYTDLPToolsReplaceMaintain(ctx context.Context) (bool, error) {
	v, err := uc.repo.Get(ctx, keyYTDLPToolsReplaceMaintain)
	if err != nil {
		return false, err
	}
	return parseBoolSetting(v), nil
}

// SetYTDLPToolsReplaceMaintain persists the maintain desired toggle.
func (uc *SettingsUseCase) SetYTDLPToolsReplaceMaintain(ctx context.Context, on bool) error {
	return uc.repo.Set(ctx, keyYTDLPToolsReplaceMaintain, strconv.FormatBool(on))
}

// GetYTDLPToolsReplaceRiskAck returns whether Tools replace risk acknowledgment was recorded.
func (uc *SettingsUseCase) GetYTDLPToolsReplaceRiskAck(ctx context.Context) (bool, error) {
	v, err := uc.repo.Get(ctx, keyYTDLPToolsReplaceRiskAck)
	if err != nil {
		return false, err
	}
	return parseBoolSetting(v), nil
}

// SetYTDLPToolsReplaceRiskAck persists risk acknowledgment.
func (uc *SettingsUseCase) SetYTDLPToolsReplaceRiskAck(ctx context.Context, ack bool) error {
	return uc.repo.Set(ctx, keyYTDLPToolsReplaceRiskAck, strconv.FormatBool(ack))
}

// GetYTDLPToolsReplacePendingError returns the last pending re-link error message (empty if none).
func (uc *SettingsUseCase) GetYTDLPToolsReplacePendingError(ctx context.Context) (string, error) {
	return uc.repo.Get(ctx, keyYTDLPToolsReplacePendingError)
}

// SetYTDLPToolsReplacePendingError persists or clears the pending re-link error.
func (uc *SettingsUseCase) SetYTDLPToolsReplacePendingError(ctx context.Context, msg string) error {
	return uc.repo.Set(ctx, keyYTDLPToolsReplacePendingError, strings.TrimSpace(msg))
}

// GetYTDLPOfficialCacheTag returns the last recorded official cache release tag (empty if unset).
func (uc *SettingsUseCase) GetYTDLPOfficialCacheTag(ctx context.Context) (string, error) {
	return uc.repo.Get(ctx, keyYTDLPOfficialCacheTag)
}

// SetYTDLPOfficialCacheTag records the release tag installed into the official cache.
func (uc *SettingsUseCase) SetYTDLPOfficialCacheTag(ctx context.Context, tag string) error {
	return uc.repo.Set(ctx, keyYTDLPOfficialCacheTag, normalizeReleaseTag(tag))
}

// GetYTDLPKnownLatest returns the last GitHub latest release metadata shown in the Video tab.
// Missing keys are ignored (same pattern as GetPathSettings); partial results are returned.
func (uc *SettingsUseCase) GetYTDLPKnownLatest(ctx context.Context) (version, tag, downloadURL string) {
	version, _ = uc.repo.Get(ctx, keyYTDLPKnownLatestVersion)
	tag, _ = uc.repo.Get(ctx, keyYTDLPKnownLatestTag)
	downloadURL, _ = uc.repo.Get(ctx, keyYTDLPKnownLatestDownloadURL)
	return version, tag, downloadURL
}

// SetYTDLPKnownLatest persists GitHub latest release metadata for the Video tab.
func (uc *SettingsUseCase) SetYTDLPKnownLatest(ctx context.Context, version, tag, downloadURL string) error {
	if err := uc.repo.Set(ctx, keyYTDLPKnownLatestVersion, normalizeReleaseTag(version)); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keyYTDLPKnownLatestTag, normalizeReleaseTag(tag)); err != nil {
		return err
	}
	return uc.repo.Set(ctx, keyYTDLPKnownLatestDownloadURL, strings.TrimSpace(downloadURL))
}

func parseBoolSetting(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// PathSettings holds VRChat/Steam/output_log paths (Wails-bound; camelCase json tags).
type PathSettings struct {
	VRChatPathWindows string `json:"vrchatPathWindows"`
	SteamPathLinux    string `json:"steamPathLinux"`
	OutputLogPath     string `json:"outputLogPath"`
}

// GetPathSettings returns all path settings.
func (uc *SettingsUseCase) GetPathSettings(ctx context.Context) (*PathSettings, error) {
	vrchat, _ := uc.repo.Get(ctx, keyVRChatPathWindows)
	steam, _ := uc.repo.Get(ctx, keySteamPathLinux)
	outputLog, _ := uc.repo.Get(ctx, keyOutputLogPath)
	return &PathSettings{
		VRChatPathWindows: vrchat,
		SteamPathLinux:    steam,
		OutputLogPath:     outputLog,
	}, nil
}

// SetPathSettings saves all path settings.
func (uc *SettingsUseCase) SetPathSettings(ctx context.Context, ps *PathSettings) error {
	if ps == nil {
		return nil
	}
	vrchat := strings.TrimSpace(ps.VRChatPathWindows)
	steam := strings.TrimSpace(ps.SteamPathLinux)
	outputLog := strings.TrimSpace(ps.OutputLogPath)
	if err := validateOutputLogPathSetting(outputLog); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keyVRChatPathWindows, vrchat); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keySteamPathLinux, steam); err != nil {
		return err
	}
	return uc.repo.Set(ctx, keyOutputLogPath, outputLog)
}

// ValidatePath checks if the path exists and is accessible (file or executable in PATH).
func (uc *SettingsUseCase) ValidatePath(path string) bool {
	if path == "" {
		return false
	}
	if filepath.VolumeName(path) == "" && filepath.Dir(path) == "." && filepath.Base(path) == path {
		_, err := exec.LookPath(path)
		return err == nil
	}
	info, err := os.Stat(path)
	return err == nil && info != nil && info.Mode().IsRegular()
}
