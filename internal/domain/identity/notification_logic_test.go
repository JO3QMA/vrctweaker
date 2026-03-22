package identity

import (
	"testing"
	"time"
)

func TestIsOffline(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"offline", true},
		{"Offline", true},
		{"OFFLINE", true},
		{" offline ", true},
		{"active", false},
		{"join me", false},
		{"ask me", false},
		{"busy", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsOffline(tt.status)
		if got != tt.want {
			t.Errorf("IsOffline(%q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestDetectFavoriteOnlineTransitions(t *testing.T) {
	now := time.Now()
	mk := func(id, name, status string, fav bool) *UserCache {
		return &UserCache{VRCUserID: id, DisplayName: name, Status: status, IsFavorite: fav, LastUpdated: now}
	}

	tests := []struct {
		name   string
		before map[string]string
		after  map[string]*UserCache
		want   []string // want VRCUserIDs (order may vary)
	}{
		{
			name:   "offline_to_online",
			before: map[string]string{"u1": "offline"},
			after:  map[string]*UserCache{"u1": mk("u1", "Alice", "active", true)},
			want:   []string{"u1"},
		},
		{
			name:   "offline_to_offline_no_notify",
			before: map[string]string{"u1": "offline"},
			after:  map[string]*UserCache{"u1": mk("u1", "Alice", "offline", true)},
			want:   nil,
		},
		{
			name:   "online_to_online_no_notify",
			before: map[string]string{"u1": "active"},
			after:  map[string]*UserCache{"u1": mk("u1", "Alice", "join me", true)},
			want:   nil,
		},
		{
			name:   "multiple_transitions",
			before: map[string]string{"u1": "offline", "u2": "offline", "u3": "active"},
			after: map[string]*UserCache{
				"u1": mk("u1", "Alice", "active", true),
				"u2": mk("u2", "Bob", "join me", true),
				"u3": mk("u3", "Charlie", "busy", true),
			},
			want: []string{"u1", "u2"},
		},
		{
			name:   "new_favorite_no_prev_state",
			before: map[string]string{},
			after:  map[string]*UserCache{"u1": mk("u1", "Alice", "active", true)},
			want:   nil, // no previous status = cannot detect offline→online
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFavoriteOnlineTransitions(tt.before, tt.after)
			gotIDs := make(map[string]bool)
			for _, fc := range got {
				gotIDs[fc.VRCUserID] = true
			}
			wantIDs := make(map[string]bool)
			for _, id := range tt.want {
				wantIDs[id] = true
			}
			if len(gotIDs) != len(wantIDs) {
				t.Errorf("got %d transitions, want %d: got=%v want=%v", len(gotIDs), len(wantIDs), gotIDs, wantIDs)
			}
			for id := range wantIDs {
				if !gotIDs[id] {
					t.Errorf("want %q in results, not found. got=%v", id, gotIDs)
				}
			}
		})
	}
}
