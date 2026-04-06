package usecase

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/settings"
)

// SettingsUseCase handles app settings.
type SettingsUseCase struct {
	repo settings.AppSettingsRepository
}

// NewSettingsUseCase creates a new SettingsUseCase.
func NewSettingsUseCase(repo settings.AppSettingsRepository) *SettingsUseCase {
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

// GetOutputLogPath returns the output_log.txt path for VRChat (empty if not set).
func (uc *SettingsUseCase) GetOutputLogPath(ctx context.Context) (string, error) {
	return uc.repo.Get(ctx, "output_log_path")
}

// SaveOutputLogPath saves the output_log.txt path.
func (uc *SettingsUseCase) SaveOutputLogPath(ctx context.Context, path string) error {
	return uc.repo.Set(ctx, "output_log_path", path)
}

// Path settings keys in app_settings.
const (
	keyVRChatPathWindows        = "vrchat_path_windows"
	keySteamPathLinux           = "steam_path_linux"
	keyOutputLogPath            = "output_log_path"
	keyGalleryLastExitAt        = "gallery_last_exit_at"
	keySuppressSleepWhileVRChat = "suppress_sleep_while_vrchat"
)

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

// PathSettings holds VRChat/Steam/output_log paths.
type PathSettings struct {
	VRChatPathWindows string
	SteamPathLinux    string
	OutputLogPath     string
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
	if err := uc.repo.Set(ctx, keyVRChatPathWindows, ps.VRChatPathWindows); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keySteamPathLinux, ps.SteamPathLinux); err != nil {
		return err
	}
	return uc.repo.Set(ctx, keyOutputLogPath, ps.OutputLogPath)
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
