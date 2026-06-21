package locale

import (
	"os"
	"testing"
)

func TestDetect_fromEnvironment(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "ja_JP.UTF-8")
	if got := Detect(); got != "ja" {
		t.Fatalf("Detect() = %q, want ja", got)
	}
}

func TestDetect_prefersLCAll(t *testing.T) {
	t.Setenv("LC_ALL", "ko_KR.UTF-8")
	t.Setenv("LC_MESSAGES", "ja_JP.UTF-8")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := Detect(); got != "ko" {
		t.Fatalf("Detect() = %q, want ko", got)
	}
}

func TestDetect_fallsBackToLCMessages(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "zh_TW.UTF-8")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := Detect(); got != "zh-TW" {
		t.Fatalf("Detect() = %q, want zh-TW", got)
	}
}

func TestMapToAppLanguage_additionalCases(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"POSIX", "en"},
		{"posix", "en"},
		{"zh-SG", "zh-CN"},
		{"zh-MO", "zh-TW"},
		{"zh-Hans", "zh-CN"},
		{"de_DE.UTF-8", "en"},
		{"ja@variant", "ja"},
	}
	for _, tt := range tests {
		if got := MapToAppLanguage(tt.in); got != tt.want {
			t.Errorf("MapToAppLanguage(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDetect_emptyEnvDefaultsEnglish(t *testing.T) {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	if got := Detect(); got != "en" {
		t.Fatalf("Detect() = %q, want en", got)
	}
}
