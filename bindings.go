package main

import (
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/domain/vrchatconfig"
	"vrchat-tweaker/internal/usecase"
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

// LaunchArgsParsedDTO is the GUI-parsed launch arguments.
type LaunchArgsParsedDTO struct {
	NoVR                        bool   `json:"noVr"`       // -no-vr (デスクトップモード)
	ScreenMode                  string `json:"screenMode"` // fullscreen|windowed|popupwindow
	ScreenWidth                 int    `json:"screenWidth"`
	ScreenHeight                int    `json:"screenHeight"`
	FPS                         int    `json:"fps"`
	SkipRegistry                bool   `json:"skipRegistry"`
	ProcessPriority             int    `json:"processPriority"`    // -2..2, -999=omit
	MainThreadPriority          int    `json:"mainThreadPriority"` // -2..2, -999=omit
	Monitor                     int    `json:"monitor"`            // -monitor N (1-based), 0=omit
	Profile                     int    `json:"profile"`            // --profile=X, -1=omit
	EnableDebugGui              bool   `json:"enableDebugGui"`
	EnableSDKLogLevels          bool   `json:"enableSDKLogLevels"`
	EnableUdonDebugLogging      bool   `json:"enableUdonDebugLogging"`
	Midi                        string `json:"midi"`
	WatchWorlds                 bool   `json:"watchWorlds"`
	WatchAvatars                bool   `json:"watchAvatars"`
	IgnoreTrackers              string `json:"ignoreTrackers"`
	VideoDecoding               string `json:"videoDecoding"` // ""|software|hardware
	DisableAMDStutterWorkaround bool   `json:"disableAMDStutterWorkaround"`
	OSC                         string `json:"osc"`
	Affinity                    string `json:"affinity"`
	EnforceWorldServerChecks    bool   `json:"enforceWorldServerChecks"`
	Custom                      string `json:"custom"`
}

func toLaunchProfileDTOs(list []*launcher.LaunchProfile) []LaunchProfileDTO {
	out := make([]LaunchProfileDTO, len(list))
	for i, p := range list {
		out[i] = LaunchProfileDTO{
			ID:        p.ID,
			Name:      p.Name,
			Arguments: p.Arguments,
			IsDefault: p.IsDefault,
		}
		if p.CreatedAt != nil {
			s := p.CreatedAt.Format(time.RFC3339)
			out[i].CreatedAt = &s
		}
		if p.UpdatedAt != nil {
			s := p.UpdatedAt.Format(time.RFC3339)
			out[i].UpdatedAt = &s
		}
	}
	return out
}

func toLaunchArgsParsedDTO(p *launcher.LaunchArgsParsed) LaunchArgsParsedDTO {
	if p == nil {
		return LaunchArgsParsedDTO{}
	}
	return LaunchArgsParsedDTO{
		NoVR:                        p.NoVR,
		ScreenMode:                  p.ScreenMode,
		ScreenWidth:                 p.ScreenWidth,
		ScreenHeight:                p.ScreenHeight,
		FPS:                         p.FPS,
		SkipRegistry:                p.SkipRegistry,
		ProcessPriority:             p.ProcessPriority,
		MainThreadPriority:          p.MainThreadPriority,
		Monitor:                     p.Monitor,
		Profile:                     p.Profile,
		EnableDebugGui:              p.EnableDebugGui,
		EnableSDKLogLevels:          p.EnableSDKLogLevels,
		EnableUdonDebugLogging:      p.EnableUdonDebugLogging,
		Midi:                        p.Midi,
		WatchWorlds:                 p.WatchWorlds,
		WatchAvatars:                p.WatchAvatars,
		IgnoreTrackers:              p.IgnoreTrackers,
		VideoDecoding:               p.VideoDecoding,
		DisableAMDStutterWorkaround: p.DisableAMDStutterWorkaround,
		OSC:                         p.OSC,
		Affinity:                    p.Affinity,
		EnforceWorldServerChecks:    p.EnforceWorldServerChecks,
		Custom:                      p.Custom,
	}
}

func fromLaunchArgsParsedDTO(d LaunchArgsParsedDTO) *launcher.LaunchArgsParsed {
	return &launcher.LaunchArgsParsed{
		NoVR:                        d.NoVR,
		ScreenMode:                  d.ScreenMode,
		ScreenWidth:                 d.ScreenWidth,
		ScreenHeight:                d.ScreenHeight,
		FPS:                         d.FPS,
		SkipRegistry:                d.SkipRegistry,
		ProcessPriority:             d.ProcessPriority,
		MainThreadPriority:          d.MainThreadPriority,
		Monitor:                     d.Monitor,
		Profile:                     d.Profile,
		EnableDebugGui:              d.EnableDebugGui,
		EnableSDKLogLevels:          d.EnableSDKLogLevels,
		EnableUdonDebugLogging:      d.EnableUdonDebugLogging,
		Midi:                        d.Midi,
		WatchWorlds:                 d.WatchWorlds,
		WatchAvatars:                d.WatchAvatars,
		IgnoreTrackers:              d.IgnoreTrackers,
		VideoDecoding:               d.VideoDecoding,
		DisableAMDStutterWorkaround: d.DisableAMDStutterWorkaround,
		OSC:                         d.OSC,
		Affinity:                    d.Affinity,
		EnforceWorldServerChecks:    d.EnforceWorldServerChecks,
		Custom:                      d.Custom,
	}
}

func toLaunchProfile(d LaunchProfileDTO) *launcher.LaunchProfile {
	p := &launcher.LaunchProfile{
		ID:        d.ID,
		Name:      d.Name,
		Arguments: d.Arguments,
		IsDefault: d.IsDefault,
	}
	if d.CreatedAt != nil {
		t, _ := time.Parse(time.RFC3339, *d.CreatedAt)
		p.CreatedAt = &t
	}
	if d.UpdatedAt != nil {
		t, _ := time.Parse(time.RFC3339, *d.UpdatedAt)
		p.UpdatedAt = &t
	}
	return p
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

// ScanProgressDTO is emitted on Wails event gallery:scan-progress during ScanScreenshotDir.
// The backend also emits gallery:screenshots-changed (no DTO) when the picture folder watcher ingests a new screenshot.
type ScanProgressDTO struct {
	Phase   string `json:"phase"`
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Item    string `json:"item,omitempty"`
}

func toScanProgressDTO(p usecase.ScanProgress) ScanProgressDTO {
	return ScanProgressDTO{
		Phase:   p.Phase,
		Current: p.Current,
		Total:   p.Total,
		Item:    p.Item,
	}
}

// GalleryScanDoneDTO is emitted on Wails event gallery:scan-done when ScanScreenshotDir finishes.
type GalleryScanDoneDTO struct {
	Count     int    `json:"count"`
	Error     string `json:"error,omitempty"`
	Cancelled bool   `json:"cancelled"`
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
		ts := s.TakenAt.Format(time.RFC3339)
		dto.TakenAt = &ts
	}
	if s.FileSizeBytes != nil {
		dto.FileSizeBytes = s.FileSizeBytes
	}
	return dto
}

func toScreenshotFilter(d ScreenshotSearchDTO) *media.ScreenshotFilter {
	f := &media.ScreenshotFilter{
		WorldID:   d.WorldID,
		WorldName: d.WorldName,
	}
	if d.DateFrom != "" {
		if t, err := time.Parse(time.RFC3339, d.DateFrom); err == nil {
			f.FromDate = &t
		} else if t, err := time.Parse("2006-01-02", d.DateFrom); err == nil {
			f.FromDate = &t
		}
	}
	if d.DateTo != "" {
		if t, err := time.Parse(time.RFC3339, d.DateTo); err == nil {
			f.ToDate = &t
		} else if t, err := time.Parse("2006-01-02", d.DateTo); err == nil {
			f.ToDate = &t
		}
	}
	return f
}

// UserEncounterDTO is the frontend-facing encounter.
type UserEncounterDTO struct {
	ID                string `json:"id"`
	VRCUserID         string `json:"vrcUserId"`
	DisplayName       string `json:"displayName"`
	Action            string `json:"action"`
	InstanceID        string `json:"instanceId"`
	WorldID           string `json:"worldId,omitempty"`
	WorldDisplayName  string `json:"worldDisplayName,omitempty"`
	UserFirstSeenAt   string `json:"userFirstSeenAt,omitempty"`
	UserLastContactAt string `json:"userLastContactAt,omitempty"`
	IsFirstEncounter  bool   `json:"isFirstEncounter"`
	EncounteredAt     string `json:"encounteredAt"`
}

func toEncounterDTOsFromContext(list []*activity.EncounterWithContext) []UserEncounterDTO {
	out := make([]UserEncounterDTO, len(list))
	for i, row := range list {
		e := row.Encounter
		dto := UserEncounterDTO{
			ID:               e.ID,
			VRCUserID:        e.VRCUserID,
			DisplayName:      e.DisplayName,
			Action:           e.Action,
			InstanceID:       e.InstanceID,
			WorldID:          e.WorldID,
			WorldDisplayName: row.WorldDisplayName,
			IsFirstEncounter: row.IsFirstEncounter,
			EncounteredAt:    e.EncounteredAt.Format(time.RFC3339),
		}
		if row.UserFirstSeenAt != nil {
			dto.UserFirstSeenAt = row.UserFirstSeenAt.Format(time.RFC3339)
		}
		if row.UserLastContactAt != nil {
			dto.UserLastContactAt = row.UserLastContactAt.Format(time.RFC3339)
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
}

func toUserCacheDTOs(list []*identity.UserCache) []UserCacheDTO {
	out := make([]UserCacheDTO, len(list))
	for i, f := range list {
		dto := UserCacheDTO{
			VRCUserID:   f.VRCUserID,
			DisplayName: f.DisplayName,
			Status:      f.Status,
			IsFavorite:  f.IsFavorite,
			LastUpdated: f.LastUpdated.Format(time.RFC3339),
		}
		if f.FirstSeenAt != nil {
			dto.FirstSeenAt = f.FirstSeenAt.Format(time.RFC3339)
		}
		if f.LastContactAt != nil {
			dto.LastContactAt = f.LastContactAt.Format(time.RFC3339)
		}
		out[i] = dto
	}
	return out
}

// LoginResultDTO is the result of a login attempt.
type LoginResultDTO struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
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

// PathSettingsDTO is the frontend-facing path settings.
type PathSettingsDTO struct {
	VRChatPathWindows string `json:"vrchatPathWindows"`
	SteamPathLinux    string `json:"steamPathLinux"`
	OutputLogPath     string `json:"outputLogPath"`
}

func toPathSettingsDTO(ps *usecase.PathSettings) PathSettingsDTO {
	if ps == nil {
		return PathSettingsDTO{}
	}
	return PathSettingsDTO{
		VRChatPathWindows: ps.VRChatPathWindows,
		SteamPathLinux:    ps.SteamPathLinux,
		OutputLogPath:     ps.OutputLogPath,
	}
}

func toPathSettings(d PathSettingsDTO) *usecase.PathSettings {
	return &usecase.PathSettings{
		VRChatPathWindows: d.VRChatPathWindows,
		SteamPathLinux:    d.SteamPathLinux,
		OutputLogPath:     d.OutputLogPath,
	}
}

// DailyPlaySecondsDTO is a single day's play time for the frontend.
type DailyPlaySecondsDTO struct {
	Date    string `json:"date"`
	Seconds int    `json:"seconds"`
}

// TopWorldDTO is world (or aggregate) stats for the frontend.
type TopWorldDTO struct {
	WorldID   string `json:"worldId"`
	WorldName string `json:"worldName,omitempty"`
	Seconds   int    `json:"seconds"`
	Sessions  int    `json:"sessions"`
}

// ActivityStatsDTO is the frontend-facing activity statistics.
type ActivityStatsDTO struct {
	DailyPlaySeconds []DailyPlaySecondsDTO `json:"dailyPlaySeconds"`
	TopWorlds        []TopWorldDTO         `json:"topWorlds"`
}

func toActivityStatsDTO(stats *activity.ActivityStats) ActivityStatsDTO {
	if stats == nil {
		return ActivityStatsDTO{DailyPlaySeconds: []DailyPlaySecondsDTO{}, TopWorlds: []TopWorldDTO{}}
	}
	daily := make([]DailyPlaySecondsDTO, len(stats.DailyPlaySeconds))
	for i, d := range stats.DailyPlaySeconds {
		daily[i] = DailyPlaySecondsDTO{Date: d.Date, Seconds: d.Seconds}
	}
	topWorlds := make([]TopWorldDTO, len(stats.TopWorlds))
	for i, t := range stats.TopWorlds {
		topWorlds[i] = TopWorldDTO{
			WorldID:   t.WorldID,
			WorldName: t.WorldName,
			Seconds:   t.Seconds,
			Sessions:  t.Sessions,
		}
	}
	return ActivityStatsDTO{DailyPlaySeconds: daily, TopWorlds: topWorlds}
}

// AutomationRuleDTO is the frontend-facing automation rule.
type AutomationRuleDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TriggerType   string `json:"triggerType"`
	ConditionJSON string `json:"conditionJson"`
	ActionType    string `json:"actionType"`
	ActionPayload string `json:"actionPayload"`
	IsEnabled     bool   `json:"isEnabled"`
}

func toAutomationRuleDTOs(list []*automation.AutomationRule) []AutomationRuleDTO {
	out := make([]AutomationRuleDTO, len(list))
	for i, r := range list {
		out[i] = AutomationRuleDTO{
			ID:            r.ID,
			Name:          r.Name,
			TriggerType:   r.TriggerType,
			ConditionJSON: r.ConditionJSON,
			ActionType:    r.ActionType,
			ActionPayload: r.ActionPayload,
			IsEnabled:     r.IsEnabled,
		}
	}
	return out
}

func toAutomationRule(d AutomationRuleDTO) *automation.AutomationRule {
	return &automation.AutomationRule{
		ID:            d.ID,
		Name:          d.Name,
		TriggerType:   d.TriggerType,
		ConditionJSON: d.ConditionJSON,
		ActionType:    d.ActionType,
		ActionPayload: d.ActionPayload,
		IsEnabled:     d.IsEnabled,
	}
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
