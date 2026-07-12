package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"vrchat-tweaker/internal/domain/launcher"
)

type fakeLaunchProfileRepo struct {
	profiles  map[string]*launcher.LaunchProfile
	defaultID string
}

func newFakeLaunchProfileRepo() *fakeLaunchProfileRepo {
	return &fakeLaunchProfileRepo{profiles: make(map[string]*launcher.LaunchProfile)}
}

func (f *fakeLaunchProfileRepo) List(_ context.Context) ([]*launcher.LaunchProfile, error) {
	out := make([]*launcher.LaunchProfile, 0, len(f.profiles))
	for _, p := range f.profiles {
		out = append(out, p)
	}
	return out, nil
}

func (f *fakeLaunchProfileRepo) GetByID(_ context.Context, id string) (*launcher.LaunchProfile, error) {
	return f.profiles[id], nil
}

func (f *fakeLaunchProfileRepo) GetDefault(_ context.Context) (*launcher.LaunchProfile, error) {
	if f.defaultID == "" {
		return nil, nil
	}
	return f.profiles[f.defaultID], nil
}

func (f *fakeLaunchProfileRepo) Save(_ context.Context, p *launcher.LaunchProfile) error {
	cp := *p
	f.profiles[p.ID] = &cp
	if p.IsDefault {
		f.defaultID = p.ID
	}
	return nil
}

func (f *fakeLaunchProfileRepo) Delete(_ context.Context, id string) error {
	delete(f.profiles, id)
	if f.defaultID == id {
		f.defaultID = ""
	}
	return nil
}

func TestLauncherUseCase_ProfileCRUD(t *testing.T) {
	ctx := context.Background()
	repo := newFakeLaunchProfileRepo()
	uc := NewLauncherUseCase(repo)

	list, err := uc.ListProfiles(ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("ListProfiles = %d err=%v", len(list), err)
	}

	p := &launcher.LaunchProfile{Name: "Desktop", Arguments: "--no-vr", IsDefault: true}
	if saveErr := uc.SaveProfile(ctx, p); saveErr != nil {
		t.Fatal(saveErr)
	}
	if p.ID == "" || p.CreatedAt == nil || p.UpdatedAt == nil {
		t.Fatalf("SaveProfile metadata = %+v", p)
	}

	got, err := uc.GetProfile(ctx, p.ID)
	if err != nil || got == nil || got.Name != "Desktop" {
		t.Fatalf("GetProfile = %+v err=%v", got, err)
	}
	def, err := uc.GetDefaultProfile(ctx)
	if err != nil || def == nil || def.ID != p.ID {
		t.Fatalf("GetDefaultProfile = %+v err=%v", def, err)
	}

	if err := uc.DeleteProfile(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.GetProfile(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
}

func TestLauncherUseCase_LaunchVRChat_profileNotFound(t *testing.T) {
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	err := uc.LaunchVRChat(context.Background(), "missing", "", "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLauncherUseCase_LaunchToWorld_requiresWorldID(t *testing.T) {
	repo := newFakeLaunchProfileRepo()
	uc := NewLauncherUseCase(repo)
	err := uc.LaunchToWorld(context.Background(), "", "  ", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty world id")
	}
}

func TestLauncherUseCase_LaunchToWorld_profileNotFound(t *testing.T) {
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	err := uc.LaunchToWorld(context.Background(), "missing", "wrld_x", "", "", "")
	if err == nil {
		t.Fatal("expected profile not found error")
	}
}

func TestLauncherUseCase_LaunchToWorld_usesDefaultProfile(t *testing.T) {
	repo := newFakeLaunchProfileRepo()
	uc := NewLauncherUseCase(repo)
	ctx := context.Background()
	p := &launcher.LaunchProfile{Name: "Default", Arguments: "--no-vr", IsDefault: true}
	_ = uc.SaveProfile(ctx, p)

	// Launch will fail on missing steam/vrchat binary, but getProfileByIDOrDefault path is exercised.
	err := uc.LaunchToWorld(ctx, "", "wrld_test", "", "/nonexistent-steam-binary-xyz", "")
	if err == nil {
		t.Fatal("expected launch error")
	}
}

func TestLauncherUseCase_LaunchWithArgs_steamNotFoundOnLinux(t *testing.T) {
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	err := uc.LaunchWithArgs(context.Background(), "--no-vr", "", "/nonexistent-steam-binary-xyz", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "steam") {
		t.Fatalf("err = %v", err)
	}
}

func TestDefaultVRChatPathWindows(t *testing.T) {
	got := defaultVRChatPathWindows()
	if got == "" || !strings.Contains(got, "launch.exe") {
		t.Fatalf("defaultVRChatPathWindows = %q", got)
	}
}

func TestLauncherUseCase_getProfileByIDOrDefault_fallback(t *testing.T) {
	repo := newFakeLaunchProfileRepo()
	uc := NewLauncherUseCase(repo)
	ctx := context.Background()
	def := &launcher.LaunchProfile{Name: "Def", Arguments: "-x", IsDefault: true}
	_ = uc.SaveProfile(ctx, def)

	got, err := uc.getProfileByIDOrDefault(ctx, "")
	if err != nil || got == nil || got.ID != def.ID {
		t.Fatalf("getProfileByIDOrDefault = %+v err=%v", got, err)
	}
}

func TestLauncherUseCase_getProfileByIDOrDefault_unknownID(t *testing.T) {
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	_, err := uc.getProfileByIDOrDefault(context.Background(), "unknown-id")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLauncherUseCase_LaunchWithArgs_fakeSteamSuccess(t *testing.T) {
	dir := t.TempDir()
	steam := filepath.Join(dir, "fake-steam")
	if err := os.WriteFile(steam, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	if err := uc.LaunchWithArgs(context.Background(), "--no-vr", "", steam, ""); err != nil {
		t.Fatalf("LaunchWithArgs: %v", err)
	}
}

func TestLauncherUseCase_launchWindowsWithArgs_missingBinary(t *testing.T) {
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	err := uc.launchWindowsWithArgs(context.Background(), []string{"--no-vr"}, filepath.Join(t.TempDir(), "missing", "launch.exe"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "vrchat not found") {
		t.Fatalf("err = %v", err)
	}
}

type errLaunchProfileRepo struct {
	fakeLaunchProfileRepo
	getByIDErr error
}

func (e *errLaunchProfileRepo) GetByID(ctx context.Context, id string) (*launcher.LaunchProfile, error) {
	if e.getByIDErr != nil {
		return nil, e.getByIDErr
	}
	return e.fakeLaunchProfileRepo.GetByID(ctx, id)
}

func TestLauncherUseCase_getProfileByIDOrDefault_getByIDError(t *testing.T) {
	repo := &errLaunchProfileRepo{getByIDErr: errors.New("db error")}
	uc := NewLauncherUseCase(repo)
	_, err := uc.getProfileByIDOrDefault(context.Background(), "p1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLauncherUseCase_LaunchLinux_permissionDenied(t *testing.T) {
	dir := t.TempDir()
	steam := filepath.Join(dir, "steam-noexec")
	if err := os.WriteFile(steam, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	err := uc.LaunchWithArgs(context.Background(), "--no-vr", "", steam, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "execute") {
		t.Fatalf("err = %v", err)
	}
}

func TestLauncherUseCase_LaunchLinux_startFailure(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad-exec")
	if err := os.WriteFile(bad, []byte(""), 0o755); err != nil {
		t.Fatal(err)
	}
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	err := uc.LaunchWithArgs(context.Background(), "--no-vr", "", bad, "")
	if err == nil {
		t.Fatal("expected start failure")
	}
}

func TestLauncherUseCase_LaunchVRChat_withProfile(t *testing.T) {
	dir := t.TempDir()
	steam := filepath.Join(dir, "fake-steam")
	_ = os.WriteFile(steam, []byte("#!/bin/sh\nexit 0\n"), 0o755)
	repo := newFakeLaunchProfileRepo()
	uc := NewLauncherUseCase(repo)
	ctx := context.Background()
	p := &launcher.LaunchProfile{Name: "P", Arguments: "--no-vr"}
	_ = uc.SaveProfile(ctx, p)
	if err := uc.LaunchVRChat(ctx, p.ID, "", steam, ""); err != nil {
		t.Fatalf("LaunchVRChat: %v", err)
	}
}

func TestLauncherUseCase_LaunchToWorld_withDefaultProfile(t *testing.T) {
	dir := t.TempDir()
	steam := filepath.Join(dir, "fake-steam")
	_ = os.WriteFile(steam, []byte("#!/bin/sh\nexit 0\n"), 0o755)
	repo := newFakeLaunchProfileRepo()
	uc := NewLauncherUseCase(repo)
	ctx := context.Background()
	def := &launcher.LaunchProfile{Name: "Default", Arguments: "--no-vr", IsDefault: true}
	_ = uc.SaveProfile(ctx, def)
	if err := uc.LaunchToWorld(ctx, "", "wrld_join_test", "", steam, ""); err != nil {
		t.Fatalf("LaunchToWorld: %v", err)
	}
}

func TestLauncherUseCase_LaunchVRChat_getProfileError(t *testing.T) {
	repo := &errLaunchProfileRepo{getByIDErr: errors.New("read failed")}
	uc := NewLauncherUseCase(repo)
	err := uc.LaunchVRChat(context.Background(), "p1", "", "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLauncherUseCase_launchWindowsWithArgs_executableScript(t *testing.T) {
	dir := t.TempDir()
	launch := filepath.Join(dir, "launch.exe")
	if err := os.WriteFile(launch, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	uc := NewLauncherUseCase(newFakeLaunchProfileRepo())
	if err := uc.launchWindowsWithArgs(context.Background(), []string{"--no-vr"}, launch); err != nil {
		t.Fatalf("launchWindowsWithArgs: %v", err)
	}
}

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
		{
			name:       "full instance key",
			baseArgs:   "--no-vr",
			worldID:    "wrld_48cf80e6-15dd-4c17-8667-c5dc01baa5cb:88577~region(jp)",
			wantSuffix: "vrchat://launch?id=wrld_48cf80e6-15dd-4c17-8667-c5dc01baa5cb:88577~region(jp)",
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
