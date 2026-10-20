package appversion

import "testing"

func TestNormalizeProductVersion(t *testing.T) {
	t.Parallel()
	// Same rules as the former usecase normalizeReleaseTag: trim space, strip one leading "v".
	cases := []struct {
		in, want string
	}{
		{in: "v2024.12.01", want: "2024.12.01"},
		{in: "  v2024.12.01  ", want: "2024.12.01"},
		{in: "2024.12.01", want: "2024.12.01"},
		{in: "v", want: ""},
		{in: "", want: ""},
		{in: "  ", want: ""},
		{in: "V1.2.3", want: "V1.2.3"},
		{in: "vv1.0.0", want: "v1.0.0"},
	}
	for _, tc := range cases {
		if got := NormalizeProductVersion(tc.in); got != tc.want {
			t.Errorf("NormalizeProductVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
