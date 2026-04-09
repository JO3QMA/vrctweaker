package locale

import "testing"

func TestMapToAppLanguage(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "en"},
		{"en_US.UTF-8", "en"},
		{"en-GB", "en"},
		{"ja_JP.UTF-8", "ja"},
		{"ja-JP", "ja"},
		{"ko_KR.UTF-8", "ko"},
		{"zh_TW.UTF-8", "zh-TW"},
		{"zh-Hant-TW", "zh-TW"},
		{"zh_CN.UTF-8", "zh-CN"},
		{"zh-Hans-CN", "zh-CN"},
		{"zh", "zh-CN"},
		{"zh-HK", "zh-TW"},
		{"C", "en"},
	}
	for _, tt := range tests {
		if got := MapToAppLanguage(tt.in); got != tt.want {
			t.Errorf("MapToAppLanguage(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
