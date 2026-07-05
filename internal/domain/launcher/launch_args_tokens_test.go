package launcher

import (
	"reflect"
	"testing"
)

func TestParseLaunchArgsTokens(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"  ", nil},
		{"-no-vr", []string{"-no-vr"}},
		{`-batchmode "value with spaces"`, []string{"-batchmode", "value with spaces"}},
		{`'single quoted' plain`, []string{"single quoted", "plain"}},
		{"-a\t-b", []string{"-a", "-b"}},
		{`unclosed"token`, []string{"token"}},
	}
	for _, tt := range tests {
		got := parseLaunchArgsTokens(tt.in)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseLaunchArgsTokens(%q) = %#v, want %#v", tt.in, got, tt.want)
		}
	}
}

func TestParseLaunchArgsForGUI_quotedCustom(t *testing.T) {
	got := ParseLaunchArgsForGUI(`-batchmode "-custom value"`)
	if got.Custom != "-batchmode -custom value" {
		t.Fatalf("Custom = %q", got.Custom)
	}
}
