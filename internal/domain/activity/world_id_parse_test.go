package activity

import "testing"

func TestWorldIDFromInstanceKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b:64190~private(usr_x)~region(jp)", "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"},
		{"wrld_abc:1", "wrld_abc"},
		{"", ""},
		{"nope", ""},
		{"wrld_only", ""},
	}
	for _, tt := range tests {
		if got := WorldIDFromInstanceKey(tt.in); got != tt.want {
			t.Errorf("WorldIDFromInstanceKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
