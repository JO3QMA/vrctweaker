package usecase

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/launcher"
)

const cacheDirName = "Cache-WindowsPlayer"

// LauncherUseCase handles launch profile and VRChat launch logic.
type LauncherUseCase struct {
	repo launcher.LaunchProfileRepository
}

// NewLauncherUseCase creates a new LauncherUseCase.
func NewLauncherUseCase(repo launcher.LaunchProfileRepository) *LauncherUseCase {
	return &LauncherUseCase{repo: repo}
}

// ListProfiles returns all launch profiles.
func (uc *LauncherUseCase) ListProfiles(ctx context.Context) ([]*launcher.LaunchProfile, error) {
	return uc.repo.List(ctx)
}

// GetProfile returns a profile by ID.
func (uc *LauncherUseCase) GetProfile(ctx context.Context, id string) (*launcher.LaunchProfile, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetDefaultProfile returns the default profile.
func (uc *LauncherUseCase) GetDefaultProfile(ctx context.Context) (*launcher.LaunchProfile, error) {
	return uc.repo.GetDefault(ctx)
}

// SaveProfile persists a profile.
func (uc *LauncherUseCase) SaveProfile(ctx context.Context, p *launcher.LaunchProfile) error {
	now := time.Now().UTC()
	if p.ID == "" {
		p.ID = uuid.New().String()
		p.CreatedAt = &now
	}
	p.UpdatedAt = &now
	return uc.repo.Save(ctx, p)
}

// DeleteProfile removes a profile.
func (uc *LauncherUseCase) DeleteProfile(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// LaunchVRChat runs VRChat with the given profile. vrchatPath and steamPath are optional overrides.
// outputLogPath is used to resolve cache dir for --clear-cache (optional).
func (uc *LauncherUseCase) LaunchVRChat(ctx context.Context, profileID string, vrchatPath, steamPath, outputLogPath string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("profile not found: %s", profileID)
	}
	return uc.LaunchWithArgs(ctx, profile.Arguments, vrchatPath, steamPath, outputLogPath)
}

// LaunchWithArgs runs VRChat with the given arguments string.
// Handles --clear-cache (deletes cache before launch, strips from args).
// Used when launching with current GUI state without saving first.
func (uc *LauncherUseCase) LaunchWithArgs(ctx context.Context, argsStr, vrchatPath, steamPath, outputLogPath string) error {
	args, err := uc.prepareLaunchArgs(ctx, parseLaunchArgs(argsStr), outputLogPath)
	if err != nil {
		return err
	}
	return uc.launchWithArgs(ctx, args, vrchatPath, steamPath)
}

// LaunchToWorld runs VRChat with the given profile and launches into the specified world.
// Uses vrchat://launch?id=<worldID> URL scheme. profileID may be empty to use default profile.
// outputLogPath is used to resolve cache dir for --clear-cache (optional).
func (uc *LauncherUseCase) LaunchToWorld(ctx context.Context, profileID, worldID string, vrchatPath, steamPath, outputLogPath string) error {
	worldID = strings.TrimSpace(worldID)
	if worldID == "" {
		return fmt.Errorf("world_id is required for Join World")
	}
	profile, err := uc.getProfileOrDefault(ctx, profileID)
	if err != nil {
		return err
	}
	baseArgs := BuildJoinWorldArgs(profile.Arguments, worldID)
	args, err := uc.prepareLaunchArgs(ctx, baseArgs, outputLogPath)
	if err != nil {
		return err
	}
	return uc.launchWithArgs(ctx, args, vrchatPath, steamPath)
}

// BuildJoinWorldArgs returns base launch args with vrchat://launch?id=<worldID> appended.
// This is a pure function for unit testing.
func BuildJoinWorldArgs(baseArgsStr string, worldID string) []string {
	base := parseLaunchArgs(baseArgsStr)
	joinURL := "vrchat://launch?id=" + strings.TrimSpace(worldID)
	return append(base, joinURL)
}

// prepareLaunchArgs handles --clear-cache: deletes cache dir if present, returns args without --clear-cache.
func (uc *LauncherUseCase) prepareLaunchArgs(ctx context.Context, args []string, outputLogPath string) ([]string, error) {
	filtered, hadClearCache := FilterClearCacheFromArgs(args)
	if !hadClearCache {
		return filtered, nil
	}
	cacheDir, err := ResolveVRCacheDir(outputLogPath)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}
	if cacheDir != "" {
		if err := clearVRCacheDir(cacheDir); err != nil {
			return nil, fmt.Errorf("clear cache: %w", err)
		}
	}
	return filtered, nil
}

func (uc *LauncherUseCase) getProfileOrDefault(ctx context.Context, profileID string) (*launcher.LaunchProfile, error) {
	if profileID != "" {
		p, err := uc.repo.GetByID(ctx, profileID)
		if err != nil {
			return nil, err
		}
		if p != nil {
			return p, nil
		}
	}
	return uc.repo.GetDefault(ctx)
}

func (uc *LauncherUseCase) launchWithArgs(ctx context.Context, args []string, vrchatPath, steamPath string) error {
	if runtime.GOOS == "linux" {
		return uc.launchLinuxWithArgs(ctx, args, steamPath)
	}
	return uc.launchWindowsWithArgs(ctx, args, vrchatPath)
}

func (uc *LauncherUseCase) launchWindowsWithArgs(ctx context.Context, args []string, vrchatPath string) error {
	if vrchatPath == "" {
		vrchatPath = defaultVRChatPathWindows()
	}
	vrchatPath = resolveVRChatPathWindows(vrchatPath)
	if _, err := os.Stat(vrchatPath); err != nil {
		return fmt.Errorf("vrchat not found at %s: %w", vrchatPath, err)
	}

	cmd := exec.CommandContext(ctx, vrchatPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func (uc *LauncherUseCase) launchLinuxWithArgs(ctx context.Context, args []string, steamPath string) error {
	if steamPath == "" {
		steamPath = "steam"
	}
	// Validate executable before launch
	if _, err := exec.LookPath(steamPath); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("steam executable not found: %s (check steam_path_linux setting)", steamPath)
		}
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("steam executable exists but has no execute permission: %s: %w", steamPath, err)
		}
		return fmt.Errorf("steam executable validation failed: %s: %w", steamPath, err)
	}
	steamArgs := append([]string{"-applaunch", "438100"}, args...)
	cmd := exec.CommandContext(ctx, steamPath, steamArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("steam executable not found: %s", steamPath)
		}
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, fs.ErrPermission) {
			return fmt.Errorf("permission denied when launching Steam: %s: %w", steamPath, err)
		}
		return fmt.Errorf("failed to launch Steam (%s): %w", steamPath, err)
	}
	return nil
}

// parseLaunchArgs parses a command-line argument string into a slice, supporting
// quoted values (single and double quotes) so arguments with spaces are preserved.
func parseLaunchArgs(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	var cur []rune
	inDouble := false
	inSingle := false
	for _, r := range s {
		switch {
		case inDouble:
			if r == '"' {
				inDouble = false
				out = append(out, string(cur))
				cur = nil
			} else {
				cur = append(cur, r)
			}
		case inSingle:
			if r == '\'' {
				inSingle = false
				out = append(out, string(cur))
				cur = nil
			} else {
				cur = append(cur, r)
			}
		case r == '"':
			inDouble = true
			cur = nil
		case r == '\'':
			inSingle = true
			cur = nil
		case unicode.IsSpace(r):
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = nil
			}
		default:
			cur = append(cur, r)
		}
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}

func defaultVRChatPathWindows() string {
	// Use launch.exe (not VRChat.exe). Running VRChat.exe directly causes offline testing mode.
	// launch.exe is the proper entry point that connects to VRChat servers.
	return "C:\\Program Files (x86)\\Steam\\steamapps\\common\\VRChat\\launch.exe"
}

// resolveVRChatPathWindows converts VRChat.exe path to launch.exe when applicable,
// so VRChat starts online instead of in offline testing mode.
func resolveVRChatPathWindows(path string) string {
	if path == "" {
		return path
	}
	const vrchatExe = "vrchat.exe"
	pathLower := strings.ToLower(path)
	if strings.HasSuffix(pathLower, vrchatExe) {
		return path[:len(path)-len(vrchatExe)] + "launch.exe"
	}
	return path
}

// ResolveVRCacheDir returns the VRChat cache directory path.
// If outputLogPath is set, uses its parent dir + Cache-WindowsPlayer.
// Otherwise uses %USERPROFILE%\AppData\LocalLow\VRChat\VRChat\Cache-WindowsPlayer (Windows only).
// Linux is not supported (returns empty string).
func ResolveVRCacheDir(outputLogPath string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", nil // Linux not supported per spec
	}
	if outputLogPath != "" {
		parent := filepath.Dir(outputLogPath)
		return filepath.Join(parent, cacheDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "AppData", "LocalLow", "VRChat", "VRChat", cacheDirName), nil
}

// FilterClearCacheFromArgs removes --clear-cache from args and returns (filtered, hadClearCache).
func FilterClearCacheFromArgs(args []string) ([]string, bool) {
	out := make([]string, 0, len(args))
	had := false
	for _, a := range args {
		if a == "--clear-cache" {
			had = true
			continue
		}
		out = append(out, a)
	}
	return out, had
}

// clearVRCacheDir deletes the VRChat cache directory.
func clearVRCacheDir(dir string) error {
	return os.RemoveAll(dir)
}
