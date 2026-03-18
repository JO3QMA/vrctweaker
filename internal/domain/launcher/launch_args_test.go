package launcher

import (
	"reflect"
	"testing"
)

func TestParseLaunchArgsForGUI(t *testing.T) {
	tests := []struct {
		name              string
		args              string
		wantVrMode        string
		wantCache         bool
		wantScreen        string
		wantCustom        string
		wantRenderBackend string
		wantAdapter       int
	}{
		{
			name:              "empty",
			args:              "",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "no-vr short",
			args:              "-no-vr",
			wantVrMode:        VrModeDesktop,
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "no-vr long",
			args:              "--no-vr",
			wantVrMode:        VrModeDesktop,
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "vr",
			args:              "-vr",
			wantVrMode:        VrModeVR,
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "clear-cache",
			args:              "--clear-cache",
			wantVrMode:        "",
			wantCache:         true,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "fullscreen on",
			args:              "-screen-fullscreen 1",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        ScreenModeFullscreen,
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "fullscreen off",
			args:              "-screen-fullscreen 0",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "custom only",
			args:              "-batchmode",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "-batchmode",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "render backend nographics",
			args:              "-batchmode -nographics",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "-batchmode",
			wantRenderBackend: RenderBackendNoGraphics,
			wantAdapter:       -1,
		},
		{
			name:              "mixed GUI and custom",
			args:              "--no-vr --clear-cache -batchmode",
			wantVrMode:        VrModeDesktop,
			wantCache:         true,
			wantScreen:        "",
			wantCustom:        "-batchmode",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "backward compat manual no-vr",
			args:              "-no-vr -screen-fullscreen 1 -custom-arg value",
			wantVrMode:        VrModeDesktop,
			wantCache:         false,
			wantScreen:        ScreenModeFullscreen,
			wantCustom:        "-custom-arg value",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "all GUI items",
			args:              "--no-vr --clear-cache -screen-fullscreen 1",
			wantVrMode:        VrModeDesktop,
			wantCache:         true,
			wantScreen:        ScreenModeFullscreen,
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "windowed",
			args:              "-windowed",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        ScreenModeWindowed,
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "popupwindow",
			args:              "-popupwindow",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        ScreenModePopupWindow,
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       -1,
		},
		{
			name:              "adapter 0",
			args:              "-adapter 0",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       0,
		},
		{
			name:              "adapter 1",
			args:              "-adapter 1",
			wantVrMode:        "",
			wantCache:         false,
			wantScreen:        "",
			wantCustom:        "",
			wantRenderBackend: "",
			wantAdapter:       1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLaunchArgsForGUI(tt.args)
			if got.VrMode != tt.wantVrMode {
				t.Errorf("ParseLaunchArgsForGUI().VrMode = %q, want %q", got.VrMode, tt.wantVrMode)
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
			if got.RenderBackend != tt.wantRenderBackend {
				t.Errorf("ParseLaunchArgsForGUI().RenderBackend = %q, want %q", got.RenderBackend, tt.wantRenderBackend)
			}
			if got.Adapter != tt.wantAdapter {
				t.Errorf("ParseLaunchArgsForGUI().Adapter = %d, want %d", got.Adapter, tt.wantAdapter)
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
			p:    &LaunchArgsParsed{Adapter: -1},
			want: "",
		},
		{
			name: "vrMode desktop only",
			p:    &LaunchArgsParsed{VrMode: VrModeDesktop, Adapter: -1},
			want: "-no-vr",
		},
		{
			name: "clearCache only",
			p:    &LaunchArgsParsed{ClearCache: true, Adapter: -1},
			want: "--clear-cache",
		},
		{
			name: "fullscreen on only",
			p:    &LaunchArgsParsed{ScreenMode: ScreenModeFullscreen, Adapter: -1},
			want: "-screen-fullscreen 1",
		},
		{
			name: "fullscreen off",
			p:    &LaunchArgsParsed{ScreenMode: "", Adapter: -1},
			want: "",
		},
		{
			name: "custom only",
			p:    &LaunchArgsParsed{Custom: "-batchmode", Adapter: -1},
			want: "-batchmode",
		},
		{
			name: "all combined",
			p:    &LaunchArgsParsed{VrMode: VrModeDesktop, ClearCache: true, ScreenMode: ScreenModeFullscreen, Custom: "-log", Adapter: -1},
			want: "-no-vr --clear-cache -screen-fullscreen 1 -log",
		},
		{
			name: "render backend d3d11",
			p:    &LaunchArgsParsed{RenderBackend: RenderBackendD3D11, Adapter: -1},
			want: "-force-d3d11",
		},
		{
			name: "render backend vulkan",
			p:    &LaunchArgsParsed{RenderBackend: RenderBackendVulkan, Adapter: -1},
			want: "-force-vulkan",
		},
		{
			name: "render backend nographics",
			p:    &LaunchArgsParsed{RenderBackend: RenderBackendNoGraphics, Adapter: -1},
			want: "-batchmode -nographics",
		},
		{
			name: "render backend nographics with custom (batchmode deduped)",
			p:    &LaunchArgsParsed{RenderBackend: RenderBackendNoGraphics, Custom: "-batchmode -other", Adapter: -1},
			want: "-batchmode -nographics -other",
		},
		{
			name: "nil safe",
			p:    nil,
			want: "",
		},
		{
			name: "adapter 0",
			p:    &LaunchArgsParsed{Adapter: 0},
			want: "-adapter 0",
		},
		{
			name: "adapter 1",
			p:    &LaunchArgsParsed{Adapter: 1},
			want: "-adapter 1",
		},
		{
			name: "detailed options",
			p: &LaunchArgsParsed{
				VrMode: VrModeVR, FPFC: true, ScreenMode: ScreenModePopupWindow,
				ScreenWidth: 1280, ScreenHeight: 720, FPS: 72,
				Safe: true, NoSplash: true, RenderBackend: RenderBackendD3D11, ProcessPriority: 2, Adapter: 1,
			},
			want: "-vr -fpfc -popupwindow -screen-width 1280 -screen-height 720 --fps=72 -safe -nosplash -force-d3d11 --process-priority=2 -adapter 1",
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
	in := "-vr -fpfc -popupwindow -screen-width 1280 -screen-height 720 --fps=72 -safe -nosplash -noaudio --skip-registry-install -force-d3d11 -log --process-priority=2 -adapter 1"
	got := ParseLaunchArgsForGUI(in)
	if got.VrMode != VrModeVR || !got.FPFC || got.ScreenMode != ScreenModePopupWindow || got.ScreenWidth != 1280 || got.ScreenHeight != 720 ||
		got.FPS != 72 || !got.Safe || !got.NoSplash || !got.NoAudio || !got.SkipRegistry ||
		got.RenderBackend != RenderBackendD3D11 || !got.Log || got.ProcessPriority != 2 || got.Adapter != 1 {
		t.Errorf("ParseLaunchArgsForGUI(detailed) = %+v, want VrMode=vr/FPFC/PopupWindow/ScreenWidth=1280/ScreenHeight=720/FPS=72/RenderBackend=d3d11/Adapter=1/...", got)
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
		{"detailed", "-vr -fpfc -popupwindow -screen-width 1280 -screen-height 720 --fps=72 -safe -nosplash -adapter 0"},
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
	// -nographics is parsed as RenderBackend, -batchmode stays in Custom
	in := "  -batchmode  -nographics  "
	got := ParseLaunchArgsForGUI(in)
	if got.Custom == "" {
		t.Error("expected non-empty Custom for -batchmode")
	}
	if got.RenderBackend != RenderBackendNoGraphics {
		t.Errorf("expected RenderBackend nographics, got %q", got.RenderBackend)
	}
	// Roundtrip: merge then parse again
	merged := MergeLaunchArgsForGUI(got)
	reparsed := ParseLaunchArgsForGUI(merged)
	if !reflect.DeepEqual(got, reparsed) {
		t.Errorf("parse-merge-parse roundtrip: got %+v, reparsed %+v", got, reparsed)
	}
}
