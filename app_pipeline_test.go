package main

import "testing"

func Test_pipelineEventUpdatesFriends(t *testing.T) {
	t.Parallel()
	cases := []struct {
		typ  string
		want bool
	}{
		{"friend-delete", true},
		{"friend-online", true},
		{"friend-active", true},
		{"user-update", false},
		{"user-location", false},
		{"unknown", false},
	}
	for _, tc := range cases {
		if got := pipelineEventUpdatesFriends(tc.typ); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.typ, got, tc.want)
		}
	}
}
