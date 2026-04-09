package locale

import "testing"

func TestNormalizeUILocale(t *testing.T) {
	t.Parallel()
	if got := NormalizeUILocale(" ja "); got != UILangJA {
		t.Fatalf("trimmed ja: got %q", got)
	}
	if got := NormalizeUILocale("ja"); got != UILangJA {
		t.Fatalf("ja: got %q", got)
	}
	if got := NormalizeUILocale("zh-CN"); got != UILangZHCN {
		t.Fatalf("zh-CN: got %q", got)
	}
	if got := NormalizeUILocale("bogus"); got != "" {
		t.Fatalf("bogus: want empty, got %q", got)
	}
}

func TestCanonicalUILanguageForSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in     string
		want   string
		wantOk bool
	}{
		{"ja", UILangJA, true},
		{"ja-JP", UILangJA, true},
		{"en-US", UILangEN, true},
		{"zh-TW", UILangZHTW, true},
		{"fr-FR", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, ok := CanonicalUILanguageForSet(tc.in)
			if ok != tc.wantOk || got != tc.want {
				t.Fatalf("CanonicalUILanguageForSet(%q) = (%q, %v), want (%q, %v)",
					tc.in, got, ok, tc.want, tc.wantOk)
			}
		})
	}
}

func TestMatchRawLocale(t *testing.T) {
	t.Parallel()
	tests := []struct {
		raw  string
		want string
	}{
		{"", UILangEN},
		{"C", UILangEN},
		{"POSIX", UILangEN},
		{"ja_JP.UTF-8", UILangJA},
		{"ja-JP", UILangJA},
		{"en_US.UTF-8", UILangEN},
		{"ko-KR", UILangKO},
		{"zh-CN", UILangZHCN},
		{"zh-TW", UILangZHTW},
		{"zh-HK", UILangZHTW},
		{"zh-Hant", UILangZHTW},
		{"zh-Hans", UILangZHCN},
		{"zh", UILangZHCN},
		{"fr-FR", UILangEN},
		{"de", UILangEN},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()
			if got := MatchRawLocale(tc.raw); got != tc.want {
				t.Fatalf("MatchRawLocale(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}
