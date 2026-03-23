package vrchatapi

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCurrentUserProfile_JSON(t *testing.T) {
	const fixture = `{
		"id": "usr_c1644b5b-3ca4-45b4-97c6-a2a0de70d469",
		"displayName": "Display One",
		"username": "user_one",
		"status": "active",
		"statusDescription": "In a world",
		"state": "offline",
		"currentAvatarThumbnailImageUrl": "https://api.vrchat.cloud/api/1/image/file_x/1/256",
		"userIcon": "",
		"profilePicOverrideThumbnail": "",
		"authToken": "must-not-map-to-exported",
		"friends": ["usr_other"]
	}`

	var got CurrentUserProfile
	if err := json.Unmarshal([]byte(fixture), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != "usr_c1644b5b-3ca4-45b4-97c6-a2a0de70d469" {
		t.Errorf("ID = %q", got.ID)
	}
	if got.DisplayName != "Display One" || got.Username != "user_one" {
		t.Errorf("DisplayName/Username = %q / %q", got.DisplayName, got.Username)
	}
	if got.Status != "active" || got.StatusDescription != "In a world" || got.State != "offline" {
		t.Errorf("status fields = %q %q %q", got.Status, got.StatusDescription, got.State)
	}
	if got.CurrentAvatarThumbnailImageURL == "" {
		t.Error("expected avatar thumbnail URL")
	}
}

func TestClient_GetCurrentUser_notAuthenticated(t *testing.T) {
	c := NewClient("")
	_, err := c.GetCurrentUser(context.Background())
	if err != ErrNotAuthenticated {
		t.Fatalf("err = %v, want ErrNotAuthenticated", err)
	}
}
