package vrchatapi

import "testing"

func TestIsWrappedBlob(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"wrapped blob", "VRCTWKV1:abc123==", true},
		{"legacy plaintext", "authcookie_xxxx", false},
		{"empty", "", false},
		{"prefix only", "VRCTWKV1:", true},
		{"other prefix", "v2:abc123", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsWrappedBlob(tc.input); got != tc.want {
				t.Errorf("IsWrappedBlob(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
