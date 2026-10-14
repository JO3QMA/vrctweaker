//go:build windows

package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vrchat-tweaker/internal/infrastructure/appupdate"
)

// ApplyUpdate downloads the zip, verifies SHA256, and schedules replacement on Windows.
func (uc *AppUpdateUseCase) ApplyUpdate(ctx context.Context, downloadURL, zipSHA256 string) error {
	downloadURL = strings.TrimSpace(downloadURL)
	zipSHA256 = strings.ToLower(strings.TrimSpace(zipSHA256))
	if downloadURL == "" || zipSHA256 == "" {
		return fmt.Errorf("%s missing download metadata", errPrefixAppUpdate)
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
	if dlErr := uc.downloader.DownloadFile(ctx, downloadURL, zipPath, zipSHA256); dlErr != nil {
		appupdate.CleanupWorkDir(work)
		return fmt.Errorf("%s download: %w", errPrefixAppUpdate, dlErr)
	}
	stagedExe, extractErr := appupdate.ExtractWindowsExe(zipPath, filepath.Join(work, "staging"))
	if extractErr != nil {
		appupdate.CleanupWorkDir(work)
		return fmt.Errorf("%s extract: %w", errPrefixAppUpdate, extractErr)
	}
	pid := appupdate.CurrentPID()
	if schedErr := appupdate.ScheduleReplaceAndRestart(exePath, stagedExe, pid); schedErr != nil {
		appupdate.CleanupWorkDir(work)
		return fmt.Errorf("%s schedule: %w", errPrefixAppUpdate, schedErr)
	}
	appupdate.SleepBeforeQuit()
	return nil
}
