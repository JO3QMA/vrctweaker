package logwatcher

import (
	"os"
	"path/filepath"
	"time"

	"vrchat-tweaker/internal/usecase"
)

// AbsLogPath returns the absolute cleaned path for a log file.
func AbsLogPath(p string) string {
	abs, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return p
	}
	return abs
}

// MatchAbsPaths reports whether two paths refer to the same file after normalization.
func MatchAbsPaths(a, b string) bool {
	aa, e1 := filepath.Abs(filepath.Clean(a))
	bb, e2 := filepath.Abs(filepath.Clean(b))
	if e1 != nil || e2 != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}

// CheckpointVRTime formats a VRChat log line timestamp for activity log checkpoints.
func CheckpointVRTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// CloseTimeForLogFile picks a finalize timestamp for a log file on VRChat exit.
func CloseTimeForLogFile(path string, fallback time.Time, fc *usecase.ActivityLogFileCheckpoint) time.Time {
	if t, err := LastVRChatLineTimeInFile(path); err == nil && !t.IsZero() {
		return t
	}
	if fc != nil && fc.VRChatLineTime != "" {
		if t, err := time.Parse(time.RFC3339, fc.VRChatLineTime); err == nil && !t.IsZero() {
			return t
		}
	}
	return fallback
}

// BootstrapLiveLogFiles marks output_log files modified within liveWindow of the newest file.
func BootstrapLiveLogFiles(paths []string, liveWindow time.Duration) map[string]bool {
	if liveWindow <= 0 {
		liveWindow = 5 * time.Second
	}
	type fileMod struct {
		path string
		mod  time.Time
	}
	var files []fileMod
	var maxMod time.Time
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		mod := info.ModTime()
		files = append(files, fileMod{path: p, mod: mod})
		if mod.After(maxMod) {
			maxMod = mod
		}
	}
	if maxMod.IsZero() {
		return nil
	}
	live := make(map[string]bool)
	for _, f := range files {
		if maxMod.Sub(f.mod) <= liveWindow {
			live[f.path] = true
		}
	}
	return live
}
