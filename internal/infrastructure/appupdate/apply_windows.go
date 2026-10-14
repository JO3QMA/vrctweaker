//go:build windows

package appupdate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ScheduleReplaceAndRestart launches a helper that waits for pid, replaces targetExe with stagedExe, and restarts.
func ScheduleReplaceAndRestart(targetExe, stagedExe string, pid int) error {
	targetExe, err := filepath.Abs(targetExe)
	if err != nil {
		return err
	}
	stagedExe, err = filepath.Abs(stagedExe)
	if err != nil {
		return err
	}
	if pid <= 0 {
		return fmt.Errorf("appupdate: invalid pid")
	}
	installDir := filepath.Dir(targetExe)
	backupExe := targetExe + ".bak"
	scriptPath := filepath.Join(installDir, ".vrchat-tweaker-update.ps1")
	script := buildReplaceScript(targetExe, stagedExe, backupExe, pid)
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return fmt.Errorf("appupdate: write helper script: %w", err)
	}
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", scriptPath,
	)
	cmd.Dir = installDir
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("appupdate: start helper: %w", err)
	}
	return nil
}

func buildReplaceScript(targetExe, stagedExe, backupExe string, pid int) string {
	// Paths are embedded from trusted local resolution; escape single quotes for PowerShell.
	targetQ := psQuote(targetExe)
	stagedQ := psQuote(stagedExe)
	backupQ := psQuote(backupExe)
	return strings.TrimSpace(fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$target = %s
$staged = %s
$backup = %s
$targetPid = %d
try {
  $proc = Get-Process -Id $targetPid -ErrorAction SilentlyContinue
  if ($proc) { $proc.WaitForExit(120000) }
  Start-Sleep -Seconds 1
  if (Test-Path $backup) { Remove-Item -LiteralPath $backup -Force }
  if (Test-Path $target) { Move-Item -LiteralPath $target -Destination $backup -Force }
  Copy-Item -LiteralPath $staged -Destination $target -Force
  Start-Process -FilePath $target
  if (Test-Path $backup) { Remove-Item -LiteralPath $backup -Force }
} catch {
  if ((Test-Path $backup) -and -not (Test-Path $target)) {
    Move-Item -LiteralPath $backup -Destination $target -Force
  }
  exit 1
}
`, targetQ, stagedQ, backupQ, pid)) + "\n"
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// WorkDir returns a directory under installDir for update staging artifacts.
func WorkDir(installDir string) string {
	return filepath.Join(installDir, ".vrchat-tweaker-update")
}

// CleanupWorkDir removes staging files best-effort.
func CleanupWorkDir(dir string) {
	_ = os.RemoveAll(dir)
}

// CurrentPID returns the running process id.
func CurrentPID() int {
	return os.Getpid()
}

// SleepBeforeQuit gives the helper time to attach before the process exits.
func SleepBeforeQuit() {
	time.Sleep(300 * time.Millisecond)
}
