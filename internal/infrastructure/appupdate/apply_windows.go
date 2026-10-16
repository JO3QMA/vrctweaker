//go:build windows

package appupdate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// ScheduleReplaceAndRestart launches a helper that waits for pid, replaces targetExe with stagedExe, and restarts.
func ScheduleReplaceAndRestart(workDir, targetExe, stagedExe string, pid int) error {
	targetExe, err := filepath.Abs(targetExe)
	if err != nil {
		return err
	}
	stagedExe, err = filepath.Abs(stagedExe)
	if err != nil {
		return err
	}
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return err
	}
	if pid <= 0 {
		return fmt.Errorf("appupdate: invalid pid")
	}
	backupExe := targetExe + ".bak"
	scriptPath := filepath.Join(workDir, fmt.Sprintf("replace-%d.ps1", pid))
	script := buildReplaceScript(targetExe, stagedExe, backupExe, stagedExe, workDir, scriptPath, pid)
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return fmt.Errorf("appupdate: work dir: %w", err)
	}
	scriptBytes := append(utf8BOM, []byte(script))
	if err := os.WriteFile(scriptPath, scriptBytes, 0o600); err != nil {
		return fmt.Errorf("appupdate: write helper script: %w", err)
	}
	if err := startUpdateHelperProcess(scriptPath, filepath.Dir(targetExe)); err != nil {
		return fmt.Errorf("appupdate: start helper: %w", err)
	}
	return nil
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func startUpdateHelperProcess(scriptPath, workDir string) error {
	flags := []uint32{
		windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_BREAKAWAY_FROM_JOB,
		windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
	}
	var lastErr error
	for _, creationFlags := range flags {
		cmd := exec.Command(
			"powershell.exe",
			"-NoProfile",
			"-ExecutionPolicy", "Bypass",
			"-File", scriptPath,
		)
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: creationFlags}
		cmd.Dir = workDir
		if err := cmd.Start(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func buildReplaceScript(targetExe, stagedExe, backupExe, stagedCleanup, workDir, scriptPath string, pid int) string {
	targetQ := psQuote(targetExe)
	stagedQ := psQuote(stagedExe)
	backupQ := psQuote(backupExe)
	stagedCleanupQ := psQuote(stagedCleanup)
	workQ := psQuote(workDir)
	scriptQ := psQuote(scriptPath)
	return strings.TrimSpace(fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$target = %s
$staged = %s
$backup = %s
$stagedCleanup = %s
$workDir = %s
$scriptPath = %s
$targetPid = %d
try {
  $proc = Get-Process -Id $targetPid -ErrorAction SilentlyContinue
  if ($proc) {
    if (-not $proc.WaitForExit(120000)) { throw 'target process did not exit in time' }
  }
  Start-Sleep -Seconds 1
  if (Test-Path $backup) { Remove-Item -LiteralPath $backup -Force }
  if (Test-Path $target) { Move-Item -LiteralPath $target -Destination $backup -Force }
  Copy-Item -LiteralPath $staged -Destination $target -Force
  Start-Process -FilePath $target
} catch {
  if (Test-Path $backup) {
    if (Test-Path $target) { Remove-Item -LiteralPath $target -Force }
    Move-Item -LiteralPath $backup -Destination $target -Force
  }
  exit 1
} finally {
  $zip = Join-Path $workDir 'package.zip'
  if (Test-Path $zip) { Remove-Item -LiteralPath $zip -Force -ErrorAction SilentlyContinue }
  $staging = Join-Path $workDir 'staging'
  if (Test-Path $staging) { Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue }
  if (Test-Path $stagedCleanup) { Remove-Item -LiteralPath $stagedCleanup -Force -ErrorAction SilentlyContinue }
  if (Test-Path $scriptPath) { Remove-Item -LiteralPath $scriptPath -Force -ErrorAction SilentlyContinue }
}
`, targetQ, stagedQ, backupQ, stagedCleanupQ, workQ, scriptQ, pid)) + "\n"
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
