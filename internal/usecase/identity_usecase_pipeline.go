package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

// ReconcileSocialCacheFromAPIHandled runs ReconcileSocialCacheFromAPI and applies session
// expiry handling (clears token and stored credential when appropriate).
func (uc *IdentityUseCase) ReconcileSocialCacheFromAPIHandled(ctx context.Context) error {
	return uc.handleSessionError(uc.ReconcileSocialCacheFromAPI(ctx))
}

// PipelineReconnectRestSync refreshes cache from REST before (re)connecting the Pipeline.
// Only auth failures are returned as errors so transient network issues do not kill the client.
func (uc *IdentityUseCase) PipelineReconnectRestSync(ctx context.Context) error {
	err := uc.ReconcileSocialCacheFromAPI(ctx)
	if err == nil {
		return nil
	}
	if errors.Is(err, vrchatapi.ErrSessionExpired) || errors.Is(err, vrchatapi.ErrNotAuthenticated) {
		return uc.handleSessionError(err)
	}
	log.Printf("identity: PipelineReconnectRestSync: %v", err)
	return nil
}

// ReconcileSocialCacheFromAPI refreshes the self row and friend list from the REST API.
// When not authenticated, it returns nil. Friend list failures are logged and do not fail the call
// after a successful self fetch.
func (uc *IdentityUseCase) ReconcileSocialCacheFromAPI(ctx context.Context) error {
	token := uc.apiClient.GetAuthToken()
	if token == "" {
		return nil
	}
	fp := identity.AuthTokenFingerprint(token)
	if _, err := uc.fetchAndUpsertCurrentUser(ctx, fp); err != nil {
		return err
	}
	if err := uc.RefreshFriends(ctx); err != nil {
		log.Printf("identity: ReconcileSocialCacheFromAPI: RefreshFriends: %v", err)
	}
	return nil
}

// CurrentAuthToken returns the in-memory VRChat auth token, or empty when locked.
func (uc *IdentityUseCase) CurrentAuthToken() string {
	return uc.apiClient.GetAuthToken()
}

// HandleVRChatPipelineEvent applies a decoded Pipeline event to users_cache.
func (uc *IdentityUseCase) HandleVRChatPipelineEvent(ctx context.Context, typ string, payload []byte) error {
	if uc.apiClient.GetAuthToken() == "" {
		return nil
	}
	now := time.Now()
	switch typ {
	case "friend-delete":
		return uc.pipelineFriendDelete(ctx, payload, now)
	case "friend-offline":
		return uc.pipelineFriendOffline(ctx, payload, now)
	case "friend-active":
		return uc.pipelineFriendActive(ctx, payload, now)
	case "friend-online":
		return uc.pipelineFriendOnline(ctx, payload, now)
	case "friend-location":
		return uc.pipelineFriendLocation(ctx, payload, now)
	case "friend-update", "friend-add":
		return uc.pipelineFriendUser(ctx, payload, now)
	case "user-update":
		return uc.pipelineUserUpdate(ctx, payload, now)
	case "user-location":
		return uc.pipelineUserLocation(ctx, payload, now)
	default:
		return nil
	}
}

func (uc *IdentityUseCase) selfRow(ctx context.Context) (*identity.UserCache, string, error) {
	fp := identity.AuthTokenFingerprint(uc.apiClient.GetAuthToken())
	if fp == "" {
		return nil, "", nil
	}
	row, err := uc.userCacheRepo.GetSelfBySessionFingerprint(ctx, fp)
	if err != nil {
		return nil, fp, err
	}
	return row, fp, nil
}

func (uc *IdentityUseCase) pipelineFriendDelete(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline friend-delete: %v", err)
		return nil
	}
	id := strings.TrimSpace(body.UserID)
	if id == "" {
		return nil
	}
	row, err := uc.userCacheRepo.GetByVRCUserID(ctx, id)
	if err != nil || row == nil {
		return err
	}
	row.DemoteFriendToContactAfterUnfriend(now)
	return uc.userCacheRepo.Save(ctx, row)
}

func (uc *IdentityUseCase) pipelineFriendOffline(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline friend-offline: %v", err)
		return nil
	}
	return uc.pipelineSaveFriendMerge(ctx, body.UserID, now, func(u *identity.UserCache) {
		u.MergeFromPipelineFriendOffline(now)
	})
}

func (uc *IdentityUseCase) pipelineFriendActive(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID   string `json:"userId"`
		UserIDLo string `json:"userid"`
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline friend-active: %v", err)
		return nil
	}
	id := strings.TrimSpace(body.UserID)
	if id == "" {
		id = strings.TrimSpace(body.UserIDLo)
	}
	return uc.pipelineSaveFriendMerge(ctx, id, now, func(u *identity.UserCache) {
		u.MergeFromPipelineFriendActive(now, body.Platform)
	})
}

func (uc *IdentityUseCase) pipelineFriendOnline(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID   string          `json:"userId"`
		Platform string          `json:"platform"`
		Location string          `json:"location"`
		User     json.RawMessage `json:"user"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline friend-online: %v", err)
		return nil
	}
	return uc.pipelineSaveFriendMerge(ctx, body.UserID, now, func(u *identity.UserCache) {
		// Apply embedded `user` first, then envelope presence. MergeFromAPIFriend copies
		// Location/Platform from the snapshot which may be empty, private, or stale relative
		// to the top-level friend-online fields.
		if len(bytes.TrimSpace(body.User)) > 0 {
			var f vrchatapi.Friend
			if err := json.Unmarshal(body.User, &f); err == nil && strings.TrimSpace(f.ID) != "" {
				snap := userCacheFromFriend(f, u.IsFavorite, now)
				u.MergeFromPipelineFriendUser(snap, now)
			}
		}
		u.MergeFromPipelineFriendOnline(now, body.Platform, body.Location, false)
	})
}

func (uc *IdentityUseCase) pipelineFriendLocation(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID              string `json:"userId"`
		Location            string `json:"location"`
		TravelingToLocation string `json:"travelingToLocation"`
		WorldID             string `json:"worldId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline friend-location: %v", err)
		return nil
	}
	return uc.pipelineSaveFriendMerge(ctx, body.UserID, now, func(u *identity.UserCache) {
		u.MergeFromPipelineFriendLocation(now, body.Location, body.TravelingToLocation, body.WorldID)
	})
}

func (uc *IdentityUseCase) pipelineFriendUser(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID string          `json:"userId"`
		User   json.RawMessage `json:"user"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline friend-user: %v", err)
		return nil
	}
	if len(bytes.TrimSpace(body.User)) == 0 {
		return nil
	}
	var f vrchatapi.Friend
	if err := json.Unmarshal(body.User, &f); err != nil {
		log.Printf("identity: pipeline friend-user decode: %v", err)
		return nil
	}
	if strings.TrimSpace(f.ID) == "" {
		f.ID = strings.TrimSpace(body.UserID)
	}
	if f.ID == "" {
		return nil
	}
	return uc.pipelineSaveFriendMerge(ctx, f.ID, now, func(u *identity.UserCache) {
		snap := userCacheFromFriend(f, u.IsFavorite, now)
		u.MergeFromPipelineFriendUser(snap, now)
	})
}

func (uc *IdentityUseCase) pipelineUserUpdate(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID string `json:"userId"`
		User   struct {
			DisplayName                    string `json:"displayName"`
			Status                         string `json:"status"`
			StatusDescription              string `json:"statusDescription"`
			Username                       string `json:"username"`
			CurrentAvatarThumbnailImageURL string `json:"currentAvatarThumbnailImageUrl"`
			UserIcon                       string `json:"userIcon"`
			ProfilePicOverrideThumbnail    string `json:"profilePicOverrideThumbnail"`
		} `json:"user"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline user-update: %v", err)
		return nil
	}
	self, _, err := uc.selfRow(ctx)
	if err != nil || self == nil || self.UserKind != identity.UserKindSelf {
		return err
	}
	if strings.TrimSpace(body.UserID) != self.VRCUserID {
		return nil
	}
	self.MergeFromPipelineSelfUserUpdate(now,
		body.User.DisplayName, body.User.Status, body.User.StatusDescription, body.User.Username,
		body.User.CurrentAvatarThumbnailImageURL, body.User.UserIcon, body.User.ProfilePicOverrideThumbnail,
	)
	return uc.userCacheRepo.UpsertSelf(ctx, self)
}

func (uc *IdentityUseCase) pipelineUserLocation(ctx context.Context, payload []byte, now time.Time) error {
	var body struct {
		UserID              string `json:"userId"`
		Location            string `json:"location"`
		TravelingToLocation string `json:"travelingToLocation"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		log.Printf("identity: pipeline user-location: %v", err)
		return nil
	}
	self, fp, err := uc.selfRow(ctx)
	if err != nil || self == nil || self.UserKind != identity.UserKindSelf {
		return err
	}
	if strings.TrimSpace(body.UserID) != self.VRCUserID {
		return nil
	}
	self.MergeFromPipelineSelfLocation(now, body.Location, body.TravelingToLocation)
	self.SessionFingerprint = fp
	return uc.userCacheRepo.UpsertSelf(ctx, self)
}

func (uc *IdentityUseCase) pipelineSaveFriendMerge(ctx context.Context, userID string, now time.Time, fn func(*identity.UserCache)) error {
	id := strings.TrimSpace(userID)
	if id == "" {
		return nil
	}
	row, err := uc.userCacheRepo.GetByVRCUserID(ctx, id)
	if err != nil {
		return err
	}
	if row == nil {
		row = &identity.UserCache{VRCUserID: id}
	}
	if row.UserKind == identity.UserKindSelf {
		return nil
	}
	fn(row)
	return uc.userCacheRepo.Save(ctx, row)
}
