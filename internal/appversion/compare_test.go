package appversion

import "testing"

func TestCompareProductVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"2.0.0", "1.9.9", 1},
		{"v1.2.3", "1.2.3", 0},
		{"dev", "1.0.0", -1},
		{"1.0.0", "dev", 1},
		{"0.9.0-beta", "0.9.0", -1},
		{"1.2.3+build", "1.2.3", 0},
		{"1.2.3", "1.2.3+meta", 0},
		{"1.2.x", "1.2.0", -1},
		{"1.2.3.4", "1.2.3", -1},
	}
	for _, tc := range cases {
		got := CompareProductVersion(tc.a, tc.b)
		if got != tc.want {
			t.Fatalf("CompareProductVersion(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
