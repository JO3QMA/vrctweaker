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
		return err
	}
	if !info.IsDir() {
		return ErrOutputLogPathMustBeDirectory
	}
	return nil
}

// EnsureOutputLogWatchDir returns the directory to watch for output_log*.txt.
// If the stored setting is a regular file, it is migrated once to the parent directory.
// If migration is impossible, the setting is cleared and ("", nil) is returned (caller uses default).
func (uc *SettingsUseCase) EnsureOutputLogWatchDir(ctx context.Context) (string, error) {
	p, err := uc.GetOutputLogPath(ctx)
	if err != nil {
		return "", err
	}
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	absPath, absErr := filepath.Abs(filepath.Clean(p))
	if absErr != nil {
		_ = uc.repo.Set(ctx, keyOutputLogPath, "")
		return "", nil
	}
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		_ = uc.repo.Set(ctx, keyOutputLogPath, "")
		return "", nil
	}
	if info.IsDir() {
		return absPath, nil
	}
	if !info.Mode().IsRegular() {
		_ = uc.repo.Set(ctx, keyOutputLogPath, "")
		return "", nil
	}
	parent := filepath.Dir(absPath)
	parentInfo, parentErr := os.Stat(parent)
	if parentErr != nil || !parentInfo.IsDir() {
		_ = uc.repo.Set(ctx, keyOutputLogPath, "")
		return "", nil
	}
	absParent, parentAbsErr := filepath.Abs(parent)
	if parentAbsErr != nil {
		_ = uc.repo.Set(ctx, keyOutputLogPath, "")
		return "", nil
	}
	if setErr := uc.repo.Set(ctx, keyOutputLogPath, absParent); setErr != nil {
		return "", setErr
	}
	return absParent, nil
}

// Path settings keys in app_settings.
const (
	keyVRChatPathWindows        = "vrchat_path_windows"
	keySteamPathLinux           = "steam_path_linux"
	keyOutputLogPath            = "output_log_path"
	keyGalleryLastExitAt        = "gallery_last_exit_at"
	keySuppressSleepWhileVRChat = "suppress_sleep_while_vrchat"
	keyLanguage                 = "language"
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
	if err := validateOutputLogPathSetting(ps.OutputLogPath); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keyVRChatPathWindows, ps.VRChatPathWindows); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keySteamPathLinux, ps.SteamPathLinux); err != nil {
		return err
	}
	return uc.repo.Set(ctx, keyOutputLogPath, strings.TrimSpace(ps.OutputLogPath))
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
