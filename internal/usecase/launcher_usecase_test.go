package usecase

import (
	"reflect"
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
