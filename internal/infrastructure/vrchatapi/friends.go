package vrchatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Friend is a VRChat List Friends API user object (GET /auth/user/friends).
// See https://vrchat.community/reference/get-friends
type Friend struct {
	Bio                            string   `json:"bio"`
	BioLinks                       []string `json:"bioLinks"`
	CurrentAvatarImageURL          string   `json:"currentAvatarImageUrl"`
	CurrentAvatarThumbnailImageURL string   `json:"currentAvatarThumbnailImageUrl"`
	CurrentAvatarTags              []string `json:"currentAvatarTags"`
	DeveloperType                  string   `json:"developerType"`
	DisplayName                    string   `json:"displayName"`
	FriendKey                      string   `json:"friendKey"`
	ID                             string   `json:"id"`
	IsFriend                       bool     `json:"isFriend"`
	ImageURL                       string   `json:"imageUrl"`
	LastPlatform                   string   `json:"last_platform"`
	Location                       string   `json:"location"`
	LastLogin                      string   `json:"last_login"`
	LastActivity                   string   `json:"last_activity"`
	LastMobile                     string   `json:"last_mobile"`
	Platform                       string   `json:"platform"`
	ProfilePicOverride             string   `json:"profilePicOverride"`
	ProfilePicOverrideThumbnail    string   `json:"profilePicOverrideThumbnail"`
	Status                         string   `json:"status"`
	StatusDescription              string   `json:"statusDescription"`
	Tags                           []string `json:"tags"`
	UserIcon                       string   `json:"userIcon"`
	Username                       string   `json:"username"`
	UserState                      string   `json:"state"`
}

func friendsListPath(offset, n int, offline bool) string {
	q := url.Values{}
	q.Set("offset", strconv.Itoa(offset))
	q.Set("n", strconv.Itoa(n))
	q.Set("offline", strconv.FormatBool(offline))
	return "/auth/user/friends?" + q.Encode()
}

func (c *Client) getFriendsPage(ctx context.Context, offset, n int, offline bool) ([]Friend, error) {
	path := friendsListPath(offset, n, offline)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var friends []Friend
	if err := json.Unmarshal(body, &friends); err != nil {
		return nil, fmt.Errorf("decode friends page: %w", err)
	}
	return friends, nil
}

const (
	// friendsListPageSize is the VRChat List Friends API max per request (see n parameter).
	friendsListPageSize = 100
	// friendsListMaxPagesPerOfflinePass caps each offline=false / offline=true pass (~10k friends max).
	friendsListMaxPagesPerOfflinePass = 100
)

// GetFriends fetches every friend by paging GET /auth/user/friends (online/active and offline passes).
func (c *Client) GetFriends(ctx context.Context) ([]Friend, error) {
	seen := make(map[string]struct{})
	var out []Friend
	for _, offline := range []bool{false, true} {
		offset := 0
		for page := 1; ; page++ {
			if page > friendsListMaxPagesPerOfflinePass {
				return nil, fmt.Errorf("friends list: exceeded max pages (%d) for offline=%v (possible API stuck returning full pages)",
					friendsListMaxPagesPerOfflinePass, offline)
			}
			batch, err := c.getFriendsPage(ctx, offset, friendsListPageSize, offline)
			if err != nil {
				return nil, err
			}
			for _, f := range batch {
				if f.ID == "" {
					continue
				}
				if _, dup := seen[f.ID]; dup {
					continue
				}
				seen[f.ID] = struct{}{}
				out = append(out, f)
			}
			if len(batch) < friendsListPageSize {
				break
			}
			offset += friendsListPageSize
		}
	}
	return out, nil
}

// SetFavorite sets or unsets a friend as favorite.
func (c *Client) SetFavorite(ctx context.Context, userID string, favorite bool) error {
	method := http.MethodDelete
	path := fmt.Sprintf("/users/%s/favorite", userID)
	if favorite {
		method = http.MethodPut
	}
	resp, err := c.do(ctx, method, path, nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
