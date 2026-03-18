package usecase

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestBuildJoinWorldArgs(t *testing.T) {
	tests := []struct {
		name       string
		baseArgs   string
		worldID    string
		wantSuffix string
	}{
		{
			name:       "empty base",
			baseArgs:   "",
			worldID:    "wrld_48cf80e6-15dd-4c17-8667-c5dc01baa5cb",
			wantSuffix: "vrchat://launch?id=wrld_48cf80e6-15dd-4c17-8667-c5dc01baa5cb",
		},
		{
			name:       "base args plus join",
			baseArgs:   "--no-vr",
			worldID:    "wrld_abc123",
			wantSuffix: "vrchat://launch?id=wrld_abc123",
		},
		{
			name:       "profile args preserved",
			baseArgs:   "--no-vr -batchmode -nographics",
			worldID:    "wrld_xyz",
			wantSuffix: "vrchat://launch?id=wrld_xyz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildJoinWorldArgs(tt.baseArgs, tt.worldID)
			if len(got) == 0 {
				t.Errorf("BuildJoinWorldArgs() returned empty slice")
				return
			}
			last := got[len(got)-1]
			if last != tt.wantSuffix {
				t.Errorf("BuildJoinWorldArgs() last = %q, want %q", last, tt.wantSuffix)
			}
		})
	}
}

func TestBuildJoinWorldArgs_BasePlusJoin(t *testing.T) {
	base := "--no-vr -batchmode"
	worldID := "wrld_48cf80e6-15dd-4c17-8667-c5dc01baa5cb"
	got := BuildJoinWorldArgs(base, worldID)
	want := []string{"--no-vr", "-batchmode", "vrchat://launch?id=wrld_48cf80e6-15dd-4c17-8667-c5dc01baa5cb"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildJoinWorldArgs() = %v, want %v", got, want)
	}
}

func TestParseLaunchArgs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "empty",
			in:   "",
			want: nil,
		},
		{
			name: "single arg",
			in:   "--no-vr",
			want: []string{"--no-vr"},
		},
		{
			name: "multiple args space separated",
			in:   "-arg1 value1 -arg2 value2",
			want: []string{"-arg1", "value1", "-arg2", "value2"},
		},
		{
			name: "double quoted with spaces",
			in:   `--custom "value with spaces"`,
			want: []string{"--custom", "value with spaces"},
		},
		{
			name: "single quoted with spaces",
			in:   `--path 'C:\Program Files\VR'`,
			want: []string{"--path", `C:\Program Files\VR`},
		},
		{
			name: "mixed quotes",
			in:   `-a "one" -b 'two' -c three`,
			want: []string{"-a", "one", "-b", "two", "-c", "three"},
		},
		{
			name: "quoted in middle",
			in:   `before "middle part" after`,
			want: []string{"before", "middle part", "after"},
		},
		{
			name: "tab separated",
			in:   "arg1\targ2\targ3",
			want: []string{"arg1", "arg2", "arg3"},
		},
		{
			name: "multiple spaces between args",
			in:   "arg1   arg2    arg3",
			want: []string{"arg1", "arg2", "arg3"},
		},
		{
			name: "empty quoted",
			in:   `arg ""`,
			want: []string{"arg", ""},
		},
		{
			name: "vrc launch args example",
			in:   `--no-vr -batchmode -nographics`,
			want: []string{"--no-vr", "-batchmode", "-nographics"},
		},
		{
			name: "path with spaces",
			in:   `-logFile "/home/user/.local/share/Steam/steamapps/common/VRChat/Logs/output.log"`,
			want: []string{"-logFile", "/home/user/.local/share/Steam/steamapps/common/VRChat/Logs/output.log"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLaunchArgs(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("parseLaunchArgs() len = %v, want %v (got %v)", len(got), len(tt.want), got)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseLaunchArgs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolveVRCacheDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ResolveVRCacheDir Windows path logic; skip on non-Windows")
	}
	tests := []struct {
		name          string
		outputLogPath string
		wantPrefix    string
		wantSuffix    string
	}{
		{
			name:          "from output_log_path",
			outputLogPath: `C:\Users\test\AppData\LocalLow\VRChat\VRChat\output_log.txt`,
			wantSuffix:    `VRChat\Cache-WindowsPlayer`,
		},
		{
			name:          "empty uses default",
			outputLogPath: "",
			wantSuffix:    `VRChat\VRChat\Cache-WindowsPlayer`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveVRCacheDir(tt.outputLogPath)
			if err != nil {
				t.Errorf("ResolveVRCacheDir() error = %v", err)
				return
			}
			if tt.wantSuffix != "" && !strings.HasSuffix(filepath.FromSlash(got), tt.wantSuffix) {
				t.Errorf("ResolveVRCacheDir() = %q, want suffix %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestFilterClearCacheFromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut []string
		wantHad bool
	}{
		{"none", []string{"-no-vr", "-batchmode"}, []string{"-no-vr", "-batchmode"}, false},
		{"has clear-cache", []string{"-no-vr", "--clear-cache", "-batchmode"}, []string{"-no-vr", "-batchmode"}, true},
		{"only clear-cache", []string{"--clear-cache"}, []string{}, true},
		{"empty", []string{}, []string{}, false},
		{"clear-cache first", []string{"--clear-cache", "-no-vr"}, []string{"-no-vr"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, gotHad := FilterClearCacheFromArgs(tt.args)
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("FilterClearCacheFromArgs() out = %v, want %v", gotOut, tt.wantOut)
			}
			if gotHad != tt.wantHad {
				t.Errorf("FilterClearCacheFromArgs() had = %v, want %v", gotHad, tt.wantHad)
			}
		})
	}
}

func TestClearCacheBeforeLaunch_Integration(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ResolveVRCacheDir returns empty on non-Windows; skip integration test")
	}
	// Use temp dir to verify cache deletion without touching real files
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "Cache-WindowsPlayer")
	if err := os.MkdirAll(filepath.Join(cacheDir, "sub"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Simulate output_log at .../VRChat/output_log.txt
	outputLogPath := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(outputLogPath, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Resolve: parent of output_log = dir, so cache = dir/Cache-WindowsPlayer
	got, err := ResolveVRCacheDir(outputLogPath)
	if err != nil {
		t.Fatalf("ResolveVRCacheDir: %v", err)
	}
	if got != cacheDir {
		t.Fatalf("ResolveVRCacheDir = %q, want %q", got, cacheDir)
	}
	// Delete and verify
	if err := clearVRCacheDir(cacheDir); err != nil {
		t.Fatalf("clearVRCacheDir: %v", err)
	}
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Errorf("cache dir should be deleted, got err=%v", err)
	}
}

func TestResolveVRChatPathWindows(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"VRChat.exe", `C:\Program Files\Steam\steamapps\common\VRChat\VRChat.exe`, `C:\Program Files\Steam\steamapps\common\VRChat\launch.exe`},
		{"VRChat.exe lowercase", `D:\Games\VRChat\VRChat.exe`, `D:\Games\VRChat\launch.exe`},
		{"already launch.exe", `C:\VRChat\launch.exe`, `C:\VRChat\launch.exe`},
		{"other exe", `C:\Some\other.exe`, `C:\Some\other.exe`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveVRChatPathWindows(tt.in)
			if got != tt.want {
				t.Errorf("resolveVRChatPathWindows(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
