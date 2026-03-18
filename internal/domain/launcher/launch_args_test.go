package launcher

import (
	"reflect"
	"testing"
)

func TestParseLaunchArgsForGUI(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		wantNoVR   bool
		wantScreen string
		wantCustom string
	}{
		{
			name:       "empty",
			args:       "",
			wantNoVR:   false,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "no-vr parsed as NoVR",
			args:       "-no-vr",
			wantNoVR:   true,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "clear-cache goes to custom",
			args:       "--clear-cache",
			wantNoVR:   false,
			wantScreen: "",
			wantCustom: "--clear-cache",
		},
		{
			name:       "fullscreen on",
			args:       "-screen-fullscreen 1",
			wantNoVR:   false,
			wantScreen: ScreenModeFullscreen,
			wantCustom: "",
		},
		{
			name:       "fullscreen off",
			args:       "-screen-fullscreen 0",
			wantNoVR:   false,
			wantScreen: "",
			wantCustom: "",
		},
		{
			name:       "custom only",
			args:       "-batchmode",
			wantNoVR:   false,
			wantScreen: "",
			wantCustom: "-batchmode",
		},
		{
			name:       "nographics goes to custom",
			args:       "-batchmode -nographics",
			wantNoVR:   false,
			wantScreen: "",
			wantCustom: "-batchmode -nographics",
		},
		{
			name:       "mixed no-vr screen and custom",
			args:       "--no-vr -screen-fullscreen 1 -batchmode",
			wantNoVR:   true,
			wantScreen: ScreenModeFullscreen,
			wantCustom: "-batchmode",
		},
		{
			name:       "windowed",
			args:       "-windowed",
			wantNoVR:   false,
			wantScreen: ScreenModeWindowed,
			wantCustom: "",
		},
		{
			name:       "popupwindow",
			args:       "-popupwindow",
			wantNoVR:   false,
			wantScreen: ScreenModePopupWindow,
			wantCustom: "",
		},
		{
			name:       "adapter goes to custom",
			args:       "-adapter 0",
			wantNoVR:   false,
			wantScreen: "",
			wantCustom: "-adapter 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLaunchArgsForGUI(tt.args)
			if got.NoVR != tt.wantNoVR {
				t.Errorf("ParseLaunchArgsForGUI().NoVR = %v, want %v", got.NoVR, tt.wantNoVR)
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
			p:    &LaunchArgsParsed{Profile: -1, ProcessPriority: PriorityOmit, MainThreadPriority: PriorityOmit},
			want: "",
		},
		{
			name: "no-vr only",
			p:    &LaunchArgsParsed{NoVR: true, Profile: -1, ProcessPriority: PriorityOmit, MainThreadPriority: PriorityOmit},
			want: "-no-vr",
		},
		{
			name: "fullscreen on only",
			p:    &LaunchArgsParsed{ScreenMode: ScreenModeFullscreen, Profile: -1, ProcessPriority: PriorityOmit, MainThreadPriority: PriorityOmit},
			want: "-screen-fullscreen 1",
		},
		{
			name: "fullscreen off",
			p:    &LaunchArgsParsed{ScreenMode: "", Profile: -1, ProcessPriority: PriorityOmit, MainThreadPriority: PriorityOmit},
			want: "",
		},
		{
			name: "custom only",
			p:    &LaunchArgsParsed{Custom: "-batchmode", Profile: -1, ProcessPriority: PriorityOmit, MainThreadPriority: PriorityOmit},
			want: "-batchmode",
		},
		{
			name: "all combined",
			p:    &LaunchArgsParsed{ScreenMode: ScreenModeFullscreen, Custom: "-log", Profile: -1, ProcessPriority: PriorityOmit, MainThreadPriority: PriorityOmit},
			want: "-screen-fullscreen 1 -log",
		},
		{
			name: "nil safe",
			p:    nil,
			want: "",
		},
		{
			name: "detailed options",
			p: &LaunchArgsParsed{
				ScreenMode:  ScreenModePopupWindow,
				ScreenWidth: 1280, ScreenHeight: 720, FPS: 72,
				ProcessPriority: 2, Profile: -1, MainThreadPriority: PriorityOmit,
			},
			want: "-popupwindow -screen-width 1280 -screen-height 720 --fps=72 --process-priority=2",
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
	if got.ScreenMode != ScreenModePopupWindow || got.ScreenWidth != 1280 || got.ScreenHeight != 720 ||
		got.FPS != 72 || !got.SkipRegistry || got.ProcessPriority != 2 {
		t.Errorf("ParseLaunchArgsForGUI(detailed) = %+v, want PopupWindow/ScreenWidth=1280/ScreenHeight=720/FPS=72/SkipRegistry/ProcessPriority=2", got)
	}
	// Removed options (vr, fpfc, safe, nosplash, noaudio, log, adapter, force-d3d11) go to Custom
	if got.Custom != "-vr -fpfc -safe -nosplash -noaudio -force-d3d11 -log -adapter 1" {
		t.Errorf("ParseLaunchArgsForGUI(detailed).Custom = %q, want removed args in Custom", got.Custom)
	}
}

func TestParseLaunchArgsForGUI_NewOptions(t *testing.T) {
	in := "--profile=1 --enable-debug-gui --enable-sdk-log-levels --enable-udon-debug-logging --midi=MyDevice --watch-worlds --watch-avatars --ignore-trackers=a,b --disable-hw-video-decoding --disable-amd-stutter-workaround --osc=9000:127.0.0.1:9001 --affinity=FF -monitor 2 --main-thread-priority=1 --enforce-world-server-checks"
	got := ParseLaunchArgsForGUI(in)
	if got.Profile != 1 || !got.EnableDebugGui || !got.EnableSDKLogLevels || !got.EnableUdonDebugLogging ||
		got.Midi != "MyDevice" || !got.WatchWorlds || !got.WatchAvatars || got.IgnoreTrackers != "a,b" ||
		got.VideoDecoding != VideoDecodingSoftware || !got.DisableAMDStutterWorkaround ||
		got.OSC != "9000:127.0.0.1:9001" || got.Affinity != "FF" || got.Monitor != 2 ||
		got.MainThreadPriority != 1 || !got.EnforceWorldServerChecks {
		t.Errorf("ParseLaunchArgsForGUI(new options) = %+v", got)
	}
}

func TestParseLaunchArgsForGUI_ProcessPriorityRange(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want int
	}{
		{"--process-priority=-2", -2},
		{"--process-priority=-1", -1},
		{"--process-priority=0", 0},
		{"--process-priority=1", 1},
		{"--process-priority=2", 2},
	} {
		got := ParseLaunchArgsForGUI(tt.in)
		if got.ProcessPriority != tt.want {
			t.Errorf("ParseLaunchArgsForGUI(%q).ProcessPriority = %d, want %d", tt.in, got.ProcessPriority, tt.want)
		}
	}
}

func TestMergeLaunchArgsForGUI_NewOptions(t *testing.T) {
	p := &LaunchArgsParsed{
		Profile:                     1,
		EnableDebugGui:              true,
		EnableSDKLogLevels:          true,
		EnableUdonDebugLogging:      true,
		Midi:                        "MyDevice",
		WatchWorlds:                 true,
		WatchAvatars:                true,
		IgnoreTrackers:              "a,b",
		VideoDecoding:               VideoDecodingHardware,
		DisableAMDStutterWorkaround: true,
		OSC:                         "9000:127.0.0.1:9001",
		Affinity:                    "FF",
		Monitor:                     2,
		MainThreadPriority:          1,
		EnforceWorldServerChecks:    true,
		ProcessPriority:             PriorityOmit,
	}
	got := MergeLaunchArgsForGUI(p)
	// Order follows MergeLaunchArgsForGUI output sequence
	want := "-monitor 2 --profile=1 --enable-debug-gui --enable-sdk-log-levels --enable-udon-debug-logging --midi=MyDevice --watch-worlds --watch-avatars --ignore-trackers=a,b --enable-hw-video-decoding --disable-amd-stutter-workaround --osc=9000:127.0.0.1:9001 --affinity=FF --main-thread-priority=1 --enforce-world-server-checks"
	if got != want {
		t.Errorf("MergeLaunchArgsForGUI() = %q, want %q", got, want)
	}
}

func TestParseMergeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"fullscreen", "-screen-fullscreen 1"},
		{"with custom", "-screen-fullscreen 1 -batchmode -custom value"},
		{"detailed", "-popupwindow -screen-width 1280 -screen-height 720 --fps=72"},
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
	// -batchmode -nographics are unknown to GUI, both go to Custom
	in := "  -batchmode  -nographics  "
	got := ParseLaunchArgsForGUI(in)
	if got.Custom != "-batchmode -nographics" {
		t.Errorf("expected Custom %q, got %q", "-batchmode -nographics", got.Custom)
	}
	// Roundtrip: merge then parse again
	merged := MergeLaunchArgsForGUI(got)
	reparsed := ParseLaunchArgsForGUI(merged)
	if !reflect.DeepEqual(got, reparsed) {
		t.Errorf("parse-merge-parse roundtrip: got %+v, reparsed %+v", got, reparsed)
	}
}
