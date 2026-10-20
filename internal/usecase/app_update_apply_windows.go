//go:build windows

package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"vrchat-tweaker/internal/appversion"
	"vrchat-tweaker/internal/infrastructure/appupdate"
	gh "vrchat-tweaker/internal/infrastructure/github"
)

func (uc *AppUpdateUseCase) resolveReleaseForApply(ctx context.Context, expectedVersion string) (gh.VRCTweakerRelease, error) {
	if uc.releases == nil {
		return gh.VRCTweakerRelease{}, errors.New("release client not configured")
	}
	expectedVersion = appversion.NormalizeProductVersion(expectedVersion)
	if expectedVersion == "" {
		return gh.VRCTweakerRelease{}, fmt.Errorf("%s version is empty", errPrefixAppUpdate)
	}
	rel, err := uc.releases.FetchLatestStable(ctx)
	if err != nil {
		return gh.VRCTweakerRelease{}, err
	}
	if appversion.NormalizeProductVersion(rel.Version) != expectedVersion {
		return gh.VRCTweakerRelease{}, fmt.Errorf("%s release metadata changed; check again", errPrefixAppUpdate)
	}
	return rel, nil
}

// ApplyUpdate downloads the zip for latestVersion (re-fetched from GitHub), verifies SHA256, and schedules replacement.
func (uc *AppUpdateUseCase) ApplyUpdate(ctx context.Context, latestVersion string) error {
	rel, err := uc.resolveReleaseForApply(ctx, latestVersion)
	if err != nil {
		return err
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("%s resolve executable: %w", errPrefixAppUpdate, err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("%s eval executable: %w", errPrefixAppUpdate, err)
	}
	installDir := filepath.Dir(exePath)
	work := appupdate.WorkDir(installDir)
	appupdate.CleanupWorkDir(work)
	if mkdirErr := os.MkdirAll(work, 0o700); mkdirErr != nil {
		return fmt.Errorf("%s work dir: %w", errPrefixAppUpdate, mkdirErr)
	}

	zipPath := filepath.Join(work, "package.zip")
	if dlErr := uc.downloader.DownloadFile(ctx, rel.DownloadURL, zipPath, rel.ZipSHA256); dlErr != nil {
		appupdate.CleanupWorkDir(work)
		return fmt.Errorf("%s download: %w", errPrefixAppUpdate, dlErr)
	}
	stagingDir := filepath.Join(work, "staging")
	stagedExe, extractErr := appupdate.ExtractWindowsExe(zipPath, stagingDir)
	if extractErr != nil {
		appupdate.CleanupWorkDir(work)
		return fmt.Errorf("%s extract: %w", errPrefixAppUpdate, extractErr)
	}
	pid := appupdate.CurrentPID()
	if schedErr := appupdate.ScheduleReplaceAndRestart(work, exePath, stagedExe, pid); schedErr != nil {
		appupdate.CleanupWorkDir(work)
		return fmt.Errorf("%s schedule: %w", errPrefixAppUpdate, schedErr)
	}
	appupdate.SleepBeforeQuit()
	return nil
}
