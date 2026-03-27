package vrchatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestFriendsListPath(t *testing.T) {
	got := friendsListPath(120, 100, true)
	want := "/auth/user/friends?n=100&offline=true&offset=120"
	if got != want {
		// url.Values.Encode sorts keys alphabetically
		t.Fatalf("friendsListPath = %q, want %q", got, want)
	}
}

func TestFriend_JSON_full(t *testing.T) {
	const fixture = `{
		"bio": "hello",
		"bioLinks": ["https://a.example"],
		"currentAvatarImageUrl": "https://img/full",
		"currentAvatarThumbnailImageUrl": "https://img/thumb",
		"currentAvatarTags": ["tag1"],
		"developerType": "none",
		"displayName": "DN",
		"friendKey": "fk",
		"id": "usr_x",
		"isFriend": true,
		"imageUrl": "https://img",
		"last_platform": "standalonewindows",
		"location": "wrld_1:1~hidden",
		"last_login": "2019-08-24T14:15:22Z",
		"last_activity": "2019-08-24T14:15:22Z",
		"last_mobile": "2019-08-24T14:15:22Z",
		"platform": "standalonewindows",
		"profilePicOverride": "https://ppo",
		"profilePicOverrideThumbnail": "https://ppo/t",
		"status": "active",
		"statusDescription": "busy",
		"tags": ["system_trust_basic"],
		"userIcon": "https://icon",
		"username": "uname",
		"state": "active"
	}`

	var f Friend
	if err := json.Unmarshal([]byte(fixture), &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if f.ID != "usr_x" || f.DisplayName != "DN" || f.Username != "uname" {
		t.Fatalf("ids: %+v", f)
	}
	if f.Bio != "hello" || len(f.BioLinks) != 1 || f.BioLinks[0] != "https://a.example" {
		t.Fatalf("bio: %+v", f)
	}
	if f.CurrentAvatarThumbnailImageURL != "https://img/thumb" || f.CurrentAvatarImageURL != "https://img/full" {
		t.Fatalf("avatar urls: %+v", f)
	}
	if len(f.CurrentAvatarTags) != 1 || f.CurrentAvatarTags[0] != "tag1" {
		t.Fatalf("avatar tags: %+v", f)
	}
	if f.Status != "active" || f.StatusDescription != "busy" || f.UserState != "active" {
		t.Fatalf("status: %+v", f)
	}
	if f.Location != "wrld_1:1~hidden" || f.DeveloperType != "none" {
		t.Fatalf("location/dev: %+v", f)
	}
	if len(f.Tags) != 1 {
		t.Fatalf("tags: %+v", f.Tags)
	}
}

func TestClient_GetFriends_paginationAndOfflinePasses(t *testing.T) {
	t.Parallel()
	var offlineFalseCalls, offlineTrueCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/auth/user/friends" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		switch q.Get("offline") {
		case "false":
			offlineFalseCalls++
			switch q.Get("offset") {
			case "0":
				_ = json.NewEncoder(w).Encode([]Friend{{ID: "usr_a", DisplayName: "A", Status: "active"}})
			default:
				_ = json.NewEncoder(w).Encode([]Friend{})
			}
		case "true":
			offlineTrueCalls++
			switch q.Get("offset") {
			case "0":
				_ = json.NewEncoder(w).Encode([]Friend{
					{ID: "usr_b", DisplayName: "B", Status: "offline"},
					{ID: "usr_a", DisplayName: "A2", Status: "offline"},
				})
			default:
				_ = json.NewEncoder(w).Encode([]Friend{})
			}
		default:
			http.Error(w, "missing offline", http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	list, err := c.GetFriends(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if offlineFalseCalls < 1 || offlineTrueCalls < 1 {
		t.Fatalf("calls offline=false %d true %d", offlineFalseCalls, offlineTrueCalls)
	}
	if len(list) != 2 {
		t.Fatalf("merged list len = %d, want 2: %+v", len(list), list)
	}
	ids := map[string]string{}
	for _, f := range list {
		ids[f.ID] = f.DisplayName
	}
	if ids["usr_a"] != "A" {
		t.Fatalf("dedupe usr_a: want first pass display name A, got %q", ids["usr_a"])
	}
	if ids["usr_b"] != "B" {
		t.Fatalf("usr_b: %v", ids)
	}
}

func TestClient_GetFriends_errorsWhenPagesNeverShrink(t *testing.T) {
	var nCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/auth/user/friends" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("offline") != "false" {
			_ = json.NewEncoder(w).Encode([]Friend{})
			return
		}
		nCalls++
		off, _ := strconv.Atoi(q.Get("offset"))
		batch := make([]Friend, friendsListPageSize)
		for i := range batch {
			batch[i] = Friend{
				ID:          fmt.Sprintf("usr_%d_%d", off, i),
				DisplayName: "X",
				Status:      "active",
			}
		}
		_ = json.NewEncoder(w).Encode(batch)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.GetFriends(context.Background())
	if err == nil {
		t.Fatal("expected error when API keeps returning full pages beyond max")
	}
	if !strings.Contains(err.Error(), "exceeded max pages") {
		t.Fatalf("unexpected err: %v", err)
	}
	if nCalls != friendsListMaxPagesPerOfflinePass {
		t.Fatalf("HTTP calls = %d, want %d (one pass of full pages before guard)", nCalls, friendsListMaxPagesPerOfflinePass)
	}
}
