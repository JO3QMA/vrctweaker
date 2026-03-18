package launcher

import (
	"reflect"
	"testing"
)

func TestParseLaunchArgsForGUI(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		wantNoVr   bool
		wantCache  bool
		wantScreen string
		wantCustom string
	}{
		{
			name:       "empty",
			args:       "",
			wantNoVr:   false,
			wantCache:  false,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "no-vr short",
			args:       "-no-vr",
			wantNoVr:   true,
			wantCache:  false,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "no-vr long",
			args:       "--no-vr",
			wantNoVr:   true,
			wantCache:  false,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "clear-cache",
			args:       "--clear-cache",
			wantNoVr:   false,
			wantCache:  true,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "fullscreen on",
			args:       "-screen-fullscreen 1",
			wantNoVr:   false,
			wantCache:  false,
			wantScreen: ScreenModeFullscreen,
			wantCustom: "",
		},
		{
			name:       "fullscreen off",
			args:       "-screen-fullscreen 0",
			wantNoVr:   false,
			wantCache:  false,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "custom only",
			args:       "-batchmode -nographics",
			wantNoVr:   false,
			wantCache:  false,
			wantScreen: "",
			wantCustom: "-batchmode -nographics",
		},
		{
			name:       "mixed GUI and custom",
			args:       "--no-vr --clear-cache -batchmode",
			wantNoVr:   true,
			wantCache:  true,
			wantScreen: "",
			wantCustom: "-batchmode",
		},
		{
			name:       "backward compat manual no-vr",
			args:       "-no-vr -screen-fullscreen 1 -custom-arg value",
			wantNoVr:   true,
			wantCache:  false,
			wantScreen: ScreenModeFullscreen,
			wantCustom: "-custom-arg value",
		},
		{
			name:       "all GUI items",
			args:       "--no-vr --clear-cache -screen-fullscreen 1",
			wantNoVr:   true,
			wantCache:  true,
			wantScreen: ScreenModeFullscreen,
			wantCustom: "",
		},
		{
			name:       "windowed",
			args:       "-windowed",
			wantNoVr:   false,
			wantCache:  false,
			wantScreen: ScreenModeWindowed,
			wantCustom: "",
		},
		{
			name:       "popupwindow",
			args:       "-popupwindow",
			wantNoVr:   false,
			wantCache:  false,
			wantScreen: ScreenModePopupWindow,
			wantCustom: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLaunchArgsForGUI(tt.args)
			if got.NoVR != tt.wantNoVr {
				t.Errorf("ParseLaunchArgsForGUI().NoVR = %v, want %v", got.NoVR, tt.wantNoVr)
			}
			if got.ClearCache != tt.wantCache {
				t.Errorf("ParseLaunchArgsForGUI().ClearCache = %v, want %v", got.ClearCache, tt.wantCache)
			}
			if got.ScreenMode != tt.wantScreen {
				t.Errorf("ParseLaunchArgsForGUI().ScreenMode = %q, want %q", got.ScreenMode, tt.wantScreen)
			}
			if got.Custom != tt.wantCustom {
				t.Errorf("ParseLaunchArgsForGUI().Custom = %q, want %q", got.Custom, tt.wantCustom)
			}
		})
	}
}

func TestMergeLaunchArgsForGUI(t *testing.T) {
	tests := []struct {
		name string
		p    *LaunchArgsParsed
		want string
	}{
		{
			name: "empty",
			p:    &LaunchArgsParsed{},
			want: "",
		},
		{
			name: "noVr only",
			p:    &LaunchArgsParsed{NoVR: true},
			want: "-no-vr",
		},
		{
			name: "clearCache only",
			p:    &LaunchArgsParsed{ClearCache: true},
			want: "--clear-cache",
		},
		{
			name: "fullscreen on only",
			p:    &LaunchArgsParsed{ScreenMode: ScreenModeFullscreen},
			want: "-screen-fullscreen 1",
		},
		{
			name: "fullscreen off",
			p:    &LaunchArgsParsed{ScreenMode: ""},
			want: "",
		},
		{
			name: "custom only",
			p:    &LaunchArgsParsed{Custom: "-batchmode"},
			want: "-batchmode",
		},
		{
			name: "all combined",
			p:    &LaunchArgsParsed{NoVR: true, ClearCache: true, ScreenMode: ScreenModeFullscreen, Custom: "-log"},
			want: "-no-vr --clear-cache -screen-fullscreen 1 -log",
		},
		{
			name: "nil safe",
			p:    nil,
			want: "",
		},
		{
			name: "detailed options",
			p: &LaunchArgsParsed{
				VR: true, FPFC: true, ScreenMode: ScreenModePopupWindow,
				ScreenWidth: 1280, ScreenHeight: 720, FPS: 72,
				Safe: true, NoSplash: true, ProcessPriority: 2,
			},
			want: "-vr -fpfc -popupwindow -screen-width 1280 -screen-height 720 --fps=72 -safe -nosplash --process-priority=2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeLaunchArgsForGUI(tt.p)
			if got != tt.want {
				t.Errorf("MergeLaunchArgsForGUI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLaunchArgsForGUI_Detailed(t *testing.T) {
	in := "-vr -fpfc -popupwindow -screen-width 1280 -screen-height 720 --fps=72 -safe -nosplash -noaudio --skip-registry-install -force-d3d11 -log --process-priority=2"
	got := ParseLaunchArgsForGUI(in)
	if !got.VR || !got.FPFC || got.ScreenMode != ScreenModePopupWindow || got.ScreenWidth != 1280 || got.ScreenHeight != 720 ||
		got.FPS != 72 || !got.Safe || !got.NoSplash || !got.NoAudio || !got.SkipRegistry ||
		!got.ForceD3D11 || !got.Log || got.ProcessPriority != 2 {
		t.Errorf("ParseLaunchArgsForGUI(detailed) = %+v, want VR/FPFC/PopupWindow/ScreenWidth=1280/ScreenHeight=720/FPS=72/...", got)
	}
}

func TestParseMergeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"no-vr", "-no-vr"},
		{"all gui", "--no-vr --clear-cache -screen-fullscreen 1"},
		{"with custom", "--no-vr -batchmode -custom value"},
		{"detailed", "-vr -fpfc -popupwindow -screen-width 1280 -screen-height 720 --fps=72 -safe -nosplash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseLaunchArgsForGUI(tt.in)
			merged := MergeLaunchArgsForGUI(parsed)
			reparsed := ParseLaunchArgsForGUI(merged)
			if !reflect.DeepEqual(parsed, reparsed) {
				t.Errorf("roundtrip mismatch: in=%q -> merged=%q, parsed=%+v reparsed=%+v",
					tt.in, merged, parsed, reparsed)
			}
		})
	}
}

func TestParseLaunchArgsForGUI_preservesCustomOrdering(t *testing.T) {
	// Custom args should preserve their original order and trimming
	in := "  -batchmode  -nographics  "
	got := ParseLaunchArgsForGUI(in)
	if got.Custom == "" {
		t.Error("expected non-empty Custom for custom args")
	}
	// Roundtrip: merge then parse again, custom should still be recognizable
	merged := MergeLaunchArgsForGUI(got)
	reparsed := ParseLaunchArgsForGUI(merged)
	if !reflect.DeepEqual(got, reparsed) {
		t.Errorf("parse-merge-parse roundtrip: got %+v, reparsed %+v", got, reparsed)
	}
}
