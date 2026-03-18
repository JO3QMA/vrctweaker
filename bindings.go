package main

import (
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/domain/media"
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
	NoVr            bool   `json:"noVr"`
	ClearCache      bool   `json:"clearCache"`
	ScreenMode      string `json:"screenMode"` // fullscreen|windowed|popupwindow
	VR              bool   `json:"vr"`
	FPFC            bool   `json:"fpfc"`
	ScreenWidth     int    `json:"screenWidth"`
	ScreenHeight    int    `json:"screenHeight"`
	FPS             int    `json:"fps"`
	Safe            bool   `json:"safe"`
	NoSplash        bool   `json:"noSplash"`
	NoAudio         bool   `json:"noAudio"`
	SkipRegistry    bool   `json:"skipRegistry"`
	ForceD3D11      bool   `json:"forceD3d11"`
	ForceVulkan     bool   `json:"forceVulkan"`
	Log             bool   `json:"log"`
	ProcessPriority int    `json:"processPriority"`
	Custom          string `json:"custom"`
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
		NoVr:            p.NoVR,
		ClearCache:      p.ClearCache,
		ScreenMode:      p.ScreenMode,
		VR:              p.VR,
		FPFC:            p.FPFC,
		ScreenWidth:     p.ScreenWidth,
		ScreenHeight:    p.ScreenHeight,
		FPS:             p.FPS,
		Safe:            p.Safe,
		NoSplash:        p.NoSplash,
		NoAudio:         p.NoAudio,
		SkipRegistry:    p.SkipRegistry,
		ForceD3D11:      p.ForceD3D11,
		ForceVulkan:     p.ForceVulkan,
		Log:             p.Log,
		ProcessPriority: p.ProcessPriority,
		Custom:          p.Custom,
	}
}

func fromLaunchArgsParsedDTO(d LaunchArgsParsedDTO) *launcher.LaunchArgsParsed {
	return &launcher.LaunchArgsParsed{
		NoVR:            d.NoVr,
		ClearCache:      d.ClearCache,
		ScreenMode:      d.ScreenMode,
		VR:              d.VR,
		FPFC:            d.FPFC,
		ScreenWidth:     d.ScreenWidth,
		ScreenHeight:    d.ScreenHeight,
		FPS:             d.FPS,
		Safe:            d.Safe,
		NoSplash:        d.NoSplash,
		NoAudio:         d.NoAudio,
		SkipRegistry:    d.SkipRegistry,
		ForceD3D11:      d.ForceD3D11,
		ForceVulkan:     d.ForceVulkan,
		Log:             d.Log,
		ProcessPriority: d.ProcessPriority,
		Custom:          d.Custom,
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
	ID        string  `json:"id"`
	FilePath  string  `json:"filePath"`
	WorldID   string  `json:"worldId"`
	WorldName string  `json:"worldName"`
	TakenAt   *string `json:"takenAt,omitempty"`
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
		ID:        s.ID,
		FilePath:  s.FilePath,
		WorldID:   s.WorldID,
		WorldName: s.WorldName,
	}
	if s.TakenAt != nil {
		ts := s.TakenAt.Format(time.RFC3339)
		dto.TakenAt = &ts
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
	ID            string `json:"id"`
	VRCUserID     string `json:"vrcUserId"`
	DisplayName   string `json:"displayName"`
	Action        string `json:"action"`
	InstanceID    string `json:"instanceId"`
	EncounteredAt string `json:"encounteredAt"`
}

func toEncounterDTOs(list []*activity.UserEncounter) []UserEncounterDTO {
	out := make([]UserEncounterDTO, len(list))
	for i, e := range list {
		out[i] = UserEncounterDTO{
			ID:            e.ID,
			VRCUserID:     e.VRCUserID,
			DisplayName:   e.DisplayName,
			Action:        e.Action,
			InstanceID:    e.InstanceID,
			EncounteredAt: e.EncounteredAt.Format(time.RFC3339),
		}
	}
	return out
}

// FriendCacheDTO is the frontend-facing friend cache.
type FriendCacheDTO struct {
	VRCUserID   string `json:"vrcUserId"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
	IsFavorite  bool   `json:"isFavorite"`
	LastUpdated string `json:"lastUpdated"`
}

func toFriendCacheDTOs(list []*identity.FriendCache) []FriendCacheDTO {
	out := make([]FriendCacheDTO, len(list))
	for i, f := range list {
		out[i] = FriendCacheDTO{
			VRCUserID:   f.VRCUserID,
			DisplayName: f.DisplayName,
			Status:      f.Status,
			IsFavorite:  f.IsFavorite,
			LastUpdated: f.LastUpdated.Format(time.RFC3339),
		}
	}
	return out
}

// LoginResultDTO is the result of a login attempt.
type LoginResultDTO struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
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
