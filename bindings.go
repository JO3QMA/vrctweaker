package main

import (
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/domain/vrchatconfig"
)

// LaunchProfileDTO is the frontend-facing launch profile.
type LaunchProfileDTO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Arguments string  `json:"arguments"`
	IsDefault bool    `json:"isDefault"`
	CreatedAt *string `json:"createdAt,omitempty"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
}

func toLaunchProfileDTOs(list []*launcher.LaunchProfile) []LaunchProfileDTO {
	out := make([]LaunchProfileDTO, len(list))
	for i, p := range list {
		out[i] = LaunchProfileDTO{
			ID:        p.ID,
			Name:      p.Name,
			Arguments: p.Arguments,
			IsDefault: p.IsDefault,
			CreatedAt: formatRFC3339Ptr(p.CreatedAt),
			UpdatedAt: formatRFC3339Ptr(p.UpdatedAt),
		}
	}
	return out
}

func toLaunchProfile(d LaunchProfileDTO) *launcher.LaunchProfile {
	p := &launcher.LaunchProfile{
		ID:        d.ID,
		Name:      d.Name,
		Arguments: d.Arguments,
		IsDefault: d.IsDefault,
		CreatedAt: parseRFC3339Ptr(d.CreatedAt),
		UpdatedAt: parseRFC3339Ptr(d.UpdatedAt),
	}
	return p
}

// InstanceRejoinSectionDTO is the Dashboard Instance rejoin section (no instance key in JSON).
type InstanceRejoinSectionDTO struct {
	PlaySessionID     string             `json:"playSessionId"`
	WorldDisplayName  string             `json:"worldDisplayName"`
	Profiles          []LaunchProfileDTO `json:"profiles"`
	SelectedProfileID string             `json:"selectedProfileId"`
}

// ScreenshotDTO is the frontend-facing screenshot.
type ScreenshotDTO struct {
	ID                string  `json:"id"`
	FilePath          string  `json:"filePath"`
	WorldID           string  `json:"worldId"`
	WorldName         string  `json:"worldName"`
	AuthorVRCUserID   string  `json:"authorVrcUserId,omitempty"`
	AuthorDisplayName string  `json:"authorDisplayName,omitempty"`
	TakenAt           *string `json:"takenAt,omitempty"`
	FileSizeBytes     *int64  `json:"fileSizeBytes,omitempty"`
}

// ScreenshotSearchDTO is the filter for SearchScreenshots.
type ScreenshotSearchDTO struct {
	WorldID   string `json:"worldId,omitempty"`
	WorldName string `json:"worldName,omitempty"`
	DateFrom  string `json:"dateFrom,omitempty"` // ISO date or datetime
	DateTo    string `json:"dateTo,omitempty"`   // ISO date or datetime
}

func toScreenshotDTOs(list []*media.Screenshot) []ScreenshotDTO {
	out := make([]ScreenshotDTO, len(list))
	for i, s := range list {
		out[i] = *toScreenshotDTO(s)
	}
	return out
}

func toScreenshotDTO(s *media.Screenshot) *ScreenshotDTO {
	if s == nil {
		return nil
	}
	dto := &ScreenshotDTO{
		ID:                s.ID,
		FilePath:          s.FilePath,
		WorldID:           s.WorldID,
		WorldName:         s.WorldName,
		AuthorVRCUserID:   s.AuthorVRCUserID,
		AuthorDisplayName: s.AuthorDisplayName,
	}
	if s.TakenAt != nil {
		dto.TakenAt = formatRFC3339Ptr(s.TakenAt)
	}
	if s.FileSizeBytes != nil {
		v := *s.FileSizeBytes
		dto.FileSizeBytes = &v
	}
	return dto
}

func toScreenshotFilter(d ScreenshotSearchDTO) *media.ScreenshotFilter {
	f := &media.ScreenshotFilter{
		WorldID:   d.WorldID,
		WorldName: d.WorldName,
		FromDate:  parseDateOrRFC3339(d.DateFrom),
		ToDate:    parseDateOrRFC3339(d.DateTo),
	}
	return f
}

// UserEncounterDTO is the frontend-facing encounter (one row = one stay in an instance).
type UserEncounterDTO struct {
	ID                string `json:"id"`
	VRCUserID         string `json:"vrcUserId"`
	DisplayName       string `json:"displayName"`
	InstanceID        string `json:"instanceId"`
	WorldID           string `json:"worldId,omitempty"`
	WorldDisplayName  string `json:"worldDisplayName,omitempty"`
	UserFirstSeenAt   string `json:"userFirstSeenAt,omitempty"`
	UserLastContactAt string `json:"userLastContactAt,omitempty"`
	IsFirstEncounter  bool   `json:"isFirstEncounter"`
	JoinedAt          string `json:"joinedAt"`
	LeftAt            string `json:"leftAt,omitempty"`
}

func toEncounterDTOsFromContext(list []*activity.EncounterWithContext) []UserEncounterDTO {
	out := make([]UserEncounterDTO, len(list))
	for i, row := range list {
		e := row.Encounter
		dto := UserEncounterDTO{
			ID:               e.ID,
			VRCUserID:        e.VRCUserID,
			DisplayName:      e.DisplayName,
			InstanceID:       e.InstanceID,
			WorldID:          e.WorldID,
			WorldDisplayName: row.WorldDisplayName,
			IsFirstEncounter: row.IsFirstEncounter,
			JoinedAt:         formatRFC3339(e.JoinedAt),
		}
		if e.LeftAt != nil {
			dto.LeftAt = formatRFC3339(*e.LeftAt)
		}
		if row.UserFirstSeenAt != nil {
			dto.UserFirstSeenAt = formatRFC3339(*row.UserFirstSeenAt)
		}
		if row.UserLastContactAt != nil {
			dto.UserLastContactAt = formatRFC3339(*row.UserLastContactAt)
		}
		out[i] = dto
	}
	return out
}

// UserCacheDTO is the frontend-facing users_cache row for the friends list (user_kind=friend).
type UserCacheDTO struct {
	VRCUserID     string `json:"vrcUserId"`
	DisplayName   string `json:"displayName"`
	Status        string `json:"status"`
	IsFavorite    bool   `json:"isFavorite"`
	LastUpdated   string `json:"lastUpdated"`
	FirstSeenAt   string `json:"firstSeenAt,omitempty"`
	LastContactAt string `json:"lastContactAt,omitempty"`
	// Extended List Friends API fields (see VRChat GET /auth/user/friends).
	Username                    string `json:"username,omitempty"`
	StatusDescription           string `json:"statusDescription,omitempty"`
	State                       string `json:"state,omitempty"`
	CurrentAvatarThumbnailImage string `json:"currentAvatarThumbnailImageUrl,omitempty"`
	UserIcon                    string `json:"userIcon,omitempty"`
	ProfilePicOverrideThumbnail string `json:"profilePicOverrideThumbnail,omitempty"`
	Bio                         string `json:"bio,omitempty"`
	BioLinksJSON                string `json:"bioLinksJson,omitempty"`
	CurrentAvatarImageURL       string `json:"currentAvatarImageUrl,omitempty"`
	CurrentAvatarTagsJSON       string `json:"currentAvatarTagsJson,omitempty"`
	DeveloperType               string `json:"developerType,omitempty"`
	FriendKey                   string `json:"friendKey,omitempty"`
	ImageURL                    string `json:"imageUrl,omitempty"`
	LastPlatform                string `json:"lastPlatform,omitempty"`
	Location                    string `json:"location,omitempty"`
	LastLogin                   string `json:"lastLogin,omitempty"`
	LastActivity                string `json:"lastActivity,omitempty"`
	LastMobile                  string `json:"lastMobile,omitempty"`
	Platform                    string `json:"platform,omitempty"`
	ProfilePicOverride          string `json:"profilePicOverride,omitempty"`
	TagsJSON                    string `json:"tagsJson,omitempty"`
}

// UserProfileNavigationDTO is returned by ResolveUserProfileNavigation for routing (friends vs profile view).
type UserProfileNavigationDTO struct {
	User              UserCacheDTO `json:"user"`
	OpenInFriendsView bool         `json:"openInFriendsView"`
	OpenInSelfProfile bool         `json:"openInSelfProfile"`
}

func toUserCacheDTO(f *identity.UserCache) UserCacheDTO {
	if f == nil {
		return UserCacheDTO{}
	}
	dto := UserCacheDTO{
		VRCUserID:                   f.VRCUserID,
		DisplayName:                 f.DisplayName,
		Status:                      f.Status,
		IsFavorite:                  f.IsFavorite,
		LastUpdated:                 formatRFC3339(f.LastUpdated),
		Username:                    f.Username,
		StatusDescription:           f.StatusDescription,
		State:                       f.UserState,
		CurrentAvatarThumbnailImage: f.AvatarThumbnailURL,
		UserIcon:                    f.UserIconURL,
		ProfilePicOverrideThumbnail: f.ProfilePicOverrideThumbnail,
		Bio:                         f.Bio,
		BioLinksJSON:                f.BioLinksJSON,
		CurrentAvatarImageURL:       f.CurrentAvatarImageURL,
		CurrentAvatarTagsJSON:       f.CurrentAvatarTagsJSON,
		DeveloperType:               f.DeveloperType,
		FriendKey:                   f.FriendKey,
		ImageURL:                    f.ImageURL,
		LastPlatform:                f.LastPlatform,
		Location:                    f.Location,
		LastLogin:                   f.LastLogin,
		LastActivity:                f.LastActivity,
		LastMobile:                  f.LastMobile,
		Platform:                    f.Platform,
		ProfilePicOverride:          f.ProfilePicOverride,
		TagsJSON:                    f.TagsJSON,
	}
	if f.FirstSeenAt != nil {
		dto.FirstSeenAt = formatRFC3339(*f.FirstSeenAt)
	}
	if f.LastContactAt != nil {
		dto.LastContactAt = formatRFC3339(*f.LastContactAt)
	}
	return dto
}

func toUserCacheDTOs(list []*identity.UserCache) []UserCacheDTO {
	out := make([]UserCacheDTO, len(list))
	for i, f := range list {
		out[i] = toUserCacheDTO(f)
	}
	return out
}

// LoginResultDTO is the result of a login attempt.
// PlaintextToken is a one-time token the frontend must immediately wrap with Web Crypto
// and persist via PersistWrappedCredential. It must not be stored or logged.
// Over the Wails IPC bridge the field is JSON in cleartext, but that channel is local
// to the app process (not sent over the network).
type LoginResultDTO struct {
	OK             bool   `json:"ok"`
	Error          string `json:"error,omitempty"`
	PlaintextToken string `json:"plaintextToken,omitempty"`
}

// VRChatCurrentUserDTO is non-sensitive profile fields from GET /auth/user for the settings UI.
type VRChatCurrentUserDTO struct {
	ID                             string `json:"id"`
	DisplayName                    string `json:"displayName"`
	Username                       string `json:"username"`
	Status                         string `json:"status"`
	StatusDescription              string `json:"statusDescription"`
	State                          string `json:"state"`
	CurrentAvatarThumbnailImageURL string `json:"currentAvatarThumbnailImageUrl"`
	UserIcon                       string `json:"userIcon"`
	ProfilePicOverrideThumbnail    string `json:"profilePicOverrideThumbnail"`
}

// VRChatConfigDTO is the frontend-facing VRChat config.json.
type VRChatConfigDTO struct {
	CameraResWidth           int    `json:"cameraResWidth"`
	CameraResHeight          int    `json:"cameraResHeight"`
	ScreenshotResWidth       int    `json:"screenshotResWidth"`
	ScreenshotResHeight      int    `json:"screenshotResHeight"`
	PictureOutputFolder      string `json:"pictureOutputFolder"`
	PictureOutputSplitByDate *bool  `json:"pictureOutputSplitByDate"`
	FPVSteadycamFOV          int    `json:"fpvSteadycamFov"`
	CacheDirectory           string `json:"cacheDirectory"`
	CacheSize                int    `json:"cacheSize"`
	CacheExpiryDelay         int    `json:"cacheExpiryDelay"`
	DisableRichPresence      *bool  `json:"disableRichPresence"`
}

func toVRChatConfigDTO(cfg *vrchatconfig.VRChatConfig) VRChatConfigDTO {
	if cfg == nil {
		return VRChatConfigDTO{}
	}
	return VRChatConfigDTO{
		CameraResWidth:           cfg.CameraResWidth,
		CameraResHeight:          cfg.CameraResHeight,
		ScreenshotResWidth:       cfg.ScreenshotResWidth,
		ScreenshotResHeight:      cfg.ScreenshotResHeight,
		PictureOutputFolder:      cfg.PictureOutputFolder,
		PictureOutputSplitByDate: cfg.PictureOutputSplitByDate,
		FPVSteadycamFOV:          cfg.FPVSteadycamFOV,
		CacheDirectory:           cfg.CacheDirectory,
		CacheSize:                cfg.CacheSize,
		CacheExpiryDelay:         cfg.CacheExpiryDelay,
		DisableRichPresence:      cfg.DisableRichPresence,
	}
}

func fromVRChatConfigDTO(d VRChatConfigDTO) *vrchatconfig.VRChatConfig {
	return &vrchatconfig.VRChatConfig{
		CameraResWidth:           d.CameraResWidth,
		CameraResHeight:          d.CameraResHeight,
		ScreenshotResWidth:       d.ScreenshotResWidth,
		ScreenshotResHeight:      d.ScreenshotResHeight,
		PictureOutputFolder:      d.PictureOutputFolder,
		PictureOutputSplitByDate: d.PictureOutputSplitByDate,
		FPVSteadycamFOV:          d.FPVSteadycamFOV,
		CacheDirectory:           d.CacheDirectory,
		CacheSize:                d.CacheSize,
		CacheExpiryDelay:         d.CacheExpiryDelay,
		DisableRichPresence:      d.DisableRichPresence,
	}
}
