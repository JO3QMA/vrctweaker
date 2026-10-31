package wailsapp

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/locale"
	"vrchat-tweaker/internal/usecase"
)

// GetLogRetentionDays returns the log retention setting.
func (a *App) GetLogRetentionDays() (int, error) {
	return a.settings.GetLogRetentionDays(a.ctx)
}

// SetLogRetentionDays saves the log retention setting and deletes rows that are
// now past the new cutoff (same key as Activity encounters, play sessions,
// video playback history, and friend avatar observations).
func (a *App) SetLogRetentionDays(days int) error {
	if err := a.settings.SetLogRetentionDays(a.ctx, days); err != nil {
		return err
	}
	return a.purgeActivityRetention(a.ctx)
}

// GetPathSettings returns VRChat/Steam/output_log path settings.
func (a *App) GetPathSettings() (usecase.PathSettings, error) {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return usecase.PathSettings{}, err
	}
	if ps == nil {
		return usecase.PathSettings{}, nil
	}
	return *ps, nil
}

// SetPathSettings saves path settings.
func (a *App) SetPathSettings(ps usecase.PathSettings) error {
	return a.settings.SetPathSettings(a.ctx, &ps)
}

// GetSuppressSleepWhileVRChat returns whether sleep is suppressed while VRChat.exe runs (Windows).
func (a *App) GetSuppressSleepWhileVRChat() (bool, error) {
	return a.settings.GetSuppressSleepWhileVRChat(a.ctx)
}

// SetSuppressSleepWhileVRChat enables or disables sleep suppression while VRChat.exe runs.
func (a *App) SetSuppressSleepWhileVRChat(on bool) error {
	return a.settings.SetSuppressSleepWhileVRChat(a.ctx, on)
}

// GetLanguage returns the saved UI language code (ja, en, ko, zh-TW, zh-CN), or empty if unset.
func (a *App) GetLanguage() (string, error) {
	return a.settings.GetLanguage(a.ctx)
}

// SetLanguage persists the UI language code.
func (a *App) SetLanguage(lang string) error {
	return a.settings.SetLanguage(a.ctx, lang)
}

// GetSystemLocale returns the OS-preferred UI language mapped to a supported app code.
func (a *App) GetSystemLocale() string {
	return locale.Detect()
}

// ValidatePath checks if the path exists and is accessible.
func (a *App) ValidatePath(path string) bool {
	return a.settings.ValidatePath(path)
}

// OpenFileDialog opens a native file picker and returns the selected file path.
// title: dialog title, defaultDir: initial directory (empty = default), filterPattern: e.g. "*.txt" or "*.exe" (empty = all files)
func (a *App) OpenFileDialog(title, defaultDir, filterPattern string) (string, error) {
	opts := runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	}
	if filterPattern != "" {
		opts.Filters = []runtime.FileFilter{
			{DisplayName: "Filtered", Pattern: filterPattern},
			{DisplayName: "All Files", Pattern: "*"},
		}
	}
	return runtime.OpenFileDialog(a.ctx, opts)
}

// OpenDirectoryDialog opens a native directory picker and returns the selected directory path.
// title: dialog title, defaultDir: initial directory (empty = default)
func (a *App) OpenDirectoryDialog(title, defaultDir string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
}
