package usecase

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"vrchat-tweaker/internal/appversion"
	"vrchat-tweaker/internal/infrastructure/appupdate"
	gh "vrchat-tweaker/internal/infrastructure/github"
)

const appUpdateReleaseNotesMaxRunes = 500

const (
	keyAppUpdateAutoCheck       = "app_update_auto_check"
	keyDismissedUpdateVersion   = "dismissed_update_version"
	errPrefixAppUpdate          = "app update:"
	ErrAppUpdateUnsupportedSent = errPrefixAppUpdate + " auto apply is Windows only"
)

// AppUpdateStatus is returned to the frontend for update checks.
type AppUpdateStatus struct {
	CurrentVersion     string `json:"currentVersion"`
	UpdateAvailable    bool   `json:"updateAvailable"`
	ShouldPrompt       bool   `json:"shouldPrompt"`
	LatestVersion      string `json:"latestVersion"`
	LatestTag          string `json:"latestTag"`
	ReleaseNotes       string `json:"releaseNotes"`
	DownloadURL        string `json:"downloadUrl"`
	ZipSHA256          string `json:"zipSha256"`
	ReleasePageURL     string `json:"releasePageUrl"`
	AutoApplySupported bool   `json:"autoApplySupported"`
	CheckError         string `json:"checkError,omitempty"`
}

// AppUpdateUseCase checks GitHub releases and applies Windows updates.
type AppUpdateUseCase struct {
	settings   *SettingsUseCase
	releases   *gh.ReleaseClient
	downloader *appupdate.Downloader
}

// NewAppUpdateUseCase wires dependencies with defaults.
func NewAppUpdateUseCase(settings *SettingsUseCase, releases *gh.ReleaseClient, downloader *appupdate.Downloader) *AppUpdateUseCase {
	if releases == nil {
		releases = &gh.ReleaseClient{}
	}
	if downloader == nil {
		downloader = &appupdate.Downloader{}
	}
	return &AppUpdateUseCase{
		settings:   settings,
		releases:   releases,
		downloader: downloader,
	}
}

// GetAutoCheckEnabled returns whether startup should check for updates (default true).
func (uc *AppUpdateUseCase) GetAutoCheckEnabled(ctx context.Context) (bool, error) {
	if uc.settings == nil {
		return true, nil
	}
	v, err := uc.settings.Get(ctx, keyAppUpdateAutoCheck)
	if err != nil {
		return true, err
	}
	if strings.TrimSpace(v) == "" {
		return true, nil
	}
	return parseBoolSetting(v), nil
}

// SetAutoCheckEnabled persists the startup auto-check toggle.
func (uc *AppUpdateUseCase) SetAutoCheckEnabled(ctx context.Context, enabled bool) error {
	if uc.settings == nil {
		return errors.New("settings not configured")
	}
	val := "0"
	if enabled {
		val = "1"
	}
	return uc.settings.Set(ctx, keyAppUpdateAutoCheck, val)
}

// DismissUpdate records a version the user chose to skip prompting for.
func (uc *AppUpdateUseCase) DismissUpdate(ctx context.Context, version string) error {
	if uc.settings == nil {
		return errors.New("settings not configured")
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return errors.New("version is empty")
	}
	if err := appversion.ValidateProductVersion(version); err != nil {
		return err
	}
	return uc.settings.Set(ctx, keyDismissedUpdateVersion, version)
}

func (uc *AppUpdateUseCase) dismissedVersion(ctx context.Context) (string, error) {
	if uc.settings == nil {
		return "", nil
	}
	return uc.settings.Get(ctx, keyDismissedUpdateVersion)
}

// CheckForUpdate queries GitHub and compares with the running version.
// When manual is true, ShouldPrompt is set whenever an update is available (ignores dismiss).
func (uc *AppUpdateUseCase) CheckForUpdate(ctx context.Context, manual bool) (AppUpdateStatus, error) {
	st := AppUpdateStatus{
		CurrentVersion:     appversion.Version,
		AutoApplySupported: runtime.GOOS == "windows",
	}
	if uc.releases == nil {
		return st, errors.New("release client not configured")
	}
	rel, err := uc.releases.FetchLatestStable(ctx)
	if err != nil {
		st.CheckError = err.Error()
		return st, err
	}
	st.LatestVersion = rel.Version
	st.LatestTag = rel.Tag
	st.ReleaseNotes = gh.TruncateReleaseNotes(rel.Body, appUpdateReleaseNotesMaxRunes)
	st.DownloadURL = rel.DownloadURL
	st.ZipSHA256 = rel.ZipSHA256
	st.ReleasePageURL = rel.HTMLURL

	if appversion.CompareProductVersion(st.CurrentVersion, rel.Version) >= 0 {
		return st, nil
	}
	st.UpdateAvailable = true

	if manual {
		st.ShouldPrompt = true
		return st, nil
	}
	dismissed, derr := uc.dismissedVersion(ctx)
	if derr != nil {
		return st, derr
	}
	if strings.TrimSpace(dismissed) == rel.Version {
		return st, nil
	}
	st.ShouldPrompt = true
	return st, nil
}
