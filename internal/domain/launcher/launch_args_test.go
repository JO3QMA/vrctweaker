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
		wantFull   bool
		wantCustom string
	}{
		{
			name:       "empty",
			args:       "",
			wantNoVr:   false,
			wantCache:  false,
			wantFull:   false,
			wantCustom: "",
		},
		{
			name:       "no-vr short",
			args:       "-no-vr",
			wantNoVr:   true,
			wantCache:  false,
			wantFull:   false,
			wantCustom: "",
		},
		{
			name:       "no-vr long",
			args:       "--no-vr",
			wantNoVr:   true,
			wantCache:  false,
			wantFull:   false,
			wantCustom: "",
		},
		{
			name:       "clear-cache",
			args:       "--clear-cache",
			wantNoVr:   false,
			wantCache:  true,
			wantFull:   false,
			wantCustom: "",
		},
		{
			name:       "fullscreen on",
			args:       "-screen-fullscreen 1",
			wantNoVr:   false,
			wantCache:  false,
			wantFull:   true,
			wantCustom: "",
		},
		{
			name:       "fullscreen off",
			args:       "-screen-fullscreen 0",
			wantNoVr:   false,
			wantCache:  false,
			wantFull:   false,
			wantCustom: "",
		},
		{
			name:       "custom only",
			args:       "-batchmode -nographics",
			wantNoVr:   false,
			wantCache:  false,
			wantFull:   false,
			wantCustom: "-batchmode -nographics",
		},
		{
			name:       "mixed GUI and custom",
			args:       "--no-vr --clear-cache -batchmode",
			wantNoVr:   true,
			wantCache:  true,
			wantFull:   false,
			wantCustom: "-batchmode",
		},
		{
			name:       "backward compat manual no-vr",
			args:       "-no-vr -screen-fullscreen 1 -custom-arg value",
			wantNoVr:   true,
			wantCache:  false,
			wantFull:   true,
			wantCustom: "-custom-arg value",
		},
		{
			name:       "all GUI items",
			args:       "--no-vr --clear-cache -screen-fullscreen 1",
			wantNoVr:   true,
			wantCache:  true,
			wantFull:   true,
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
			if got.Fullscreen != tt.wantFull {
				t.Errorf("ParseLaunchArgsForGUI().Fullscreen = %v, want %v", got.Fullscreen, tt.wantFull)
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
			p:    &LaunchArgsParsed{Fullscreen: true},
			want: "-screen-fullscreen 1",
		},
		{
			name: "fullscreen off",
			p:    &LaunchArgsParsed{Fullscreen: false},
			want: "",
		},
		{
			name: "custom only",
			p:    &LaunchArgsParsed{Custom: "-batchmode"},
			want: "-batchmode",
		},
		{
			name: "all combined",
			p:    &LaunchArgsParsed{NoVR: true, ClearCache: true, Fullscreen: true, Custom: "-log"},
			want: "-no-vr --clear-cache -screen-fullscreen 1 -log",
		},
		{
			name: "nil safe",
			p:    nil,
			want: "",
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

func TestParseMergeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"no-vr", "-no-vr"},
		{"all gui", "--no-vr --clear-cache -screen-fullscreen 1"},
		{"with custom", "--no-vr -batchmode -custom value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseLaunchArgsForGUI(tt.in)
			merged := MergeLaunchArgsForGUI(parsed)
			reparsed := ParseLaunchArgsForGUI(merged)
			if parsed.NoVR != reparsed.NoVR || parsed.ClearCache != reparsed.ClearCache ||
				parsed.Fullscreen != reparsed.Fullscreen || parsed.Custom != reparsed.Custom {
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
