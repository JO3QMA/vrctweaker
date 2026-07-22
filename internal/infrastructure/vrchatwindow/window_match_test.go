package vrchatwindow

import "testing"

func TestClassOrTitleLooksLikeVRChat(t *testing.T) {
	cases := []struct {
		class, title string
		want         bool
	}{
		{"UnityWndClass", "", true},
		{"UnityWndClass", "VRChat", true},
		{"Other", "VRChat", true},
		{"Other", "VRChat 2022.4.2", true},
		{"Other", "NotVRChat", false},
		{"", "", false},
		{"Splash", "Loading", false},
	}
	for _, tc := range cases {
		got := classOrTitleLooksLikeVRChat(tc.class, tc.title)
		if got != tc.want {
			t.Fatalf("class=%q title=%q: got %v want %v", tc.class, tc.title, got, tc.want)
		}
	}
}
