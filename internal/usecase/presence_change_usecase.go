package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
)

const (
	// PresenceDescriptionHistoryKey is the app_settings key for status description history.
	PresenceDescriptionHistoryKey = "presence_description_history"
	maxPresenceDescriptionHistory = 20
)

var validPresenceStatuses = map[string]struct{}{
	"join me": {},
	"active":  {},
	"ask me":  {},
	"busy":    {},
}

// PresenceChangeSection is the Dashboard presence change block read model.
type PresenceChangeSection struct {
	LoggedIn          bool
	Status            string
	StatusDescription string
	History           []string
}

// PresenceChangeApplyResult is returned after a successful presence change apply.
type PresenceChangeApplyResult struct {
	Status            string
	StatusDescription string
}

// PresenceChangeUseCase handles Dashboard presence change section logic.
type PresenceChangeUseCase struct {
	identity  *IdentityUseCase
	settings  appSettingsRepo
	historyMu sync.Mutex
}

// NewPresenceChangeUseCase creates a PresenceChangeUseCase.
func NewPresenceChangeUseCase(identity *IdentityUseCase, settings appSettingsRepo) *PresenceChangeUseCase {
	return &PresenceChangeUseCase{identity: identity, settings: settings}
}

// NormalizePresenceStatus maps VRChat status to one of the four presence colors for display.
func NormalizePresenceStatus(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	if _, ok := validPresenceStatuses[s]; ok {
		return s
	}
	log.Printf("presence: unknown status %q, falling back to 'active'", status)
	return "active"
}

func validatePresenceStatus(status string) error {
	s := strings.ToLower(strings.TrimSpace(status))
	if _, ok := validPresenceStatuses[s]; ok {
		return nil
	}
	return fmt.Errorf("invalid presence status: %q", status)
}

// GetSection returns presence change section state for the Dashboard.
func (uc *PresenceChangeUseCase) GetSection(ctx context.Context) (*PresenceChangeSection, error) {
	history, err := uc.loadHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("presence change section: load history: %w", err)
	}
	loggedIn, err := uc.identity.IsLoggedIn(ctx)
	if err != nil {
		return nil, fmt.Errorf("presence change section: login check: %w", err)
	}
	if !loggedIn {
		return &PresenceChangeSection{LoggedIn: false, History: history}, nil
	}
	self, err := uc.identity.GetSelfProfile(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("presence change section: self profile: %w", err)
	}
	return &PresenceChangeSection{
		LoggedIn:          true,
		Status:            NormalizePresenceStatus(self.Status),
		StatusDescription: self.StatusDescription,
		History:           history,
	}, nil
}

// Apply updates VRChat presence and records description history on success.
func (uc *PresenceChangeUseCase) Apply(ctx context.Context, status, description string) (*PresenceChangeApplyResult, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	description = strings.TrimSpace(description)
	if err := validatePresenceStatus(status); err != nil {
		return nil, err
	}
	if err := validateVRChatStatusDescription(description); err != nil {
		return nil, err
	}
	if err := uc.identity.SetStatusAndDescription(ctx, status, description); err != nil {
		return nil, err
	}
	if description != "" {
		if err := uc.recordHistory(ctx, description); err != nil {
			log.Printf("presence change: record history: %v", err)
		}
	}
	self, err := uc.identity.GetSelfProfile(ctx, true)
	if err != nil {
		log.Printf("presence change apply: self profile refresh failed after successful status update: %v", err)
		return &PresenceChangeApplyResult{
			Status:            NormalizePresenceStatus(status),
			StatusDescription: description,
		}, nil
	}
	return &PresenceChangeApplyResult{
		Status:            NormalizePresenceStatus(self.Status),
		StatusDescription: self.StatusDescription,
	}, nil
}

func (uc *PresenceChangeUseCase) loadHistory(ctx context.Context) ([]string, error) {
	if uc.settings == nil {
		return nil, nil
	}
	raw, err := uc.settings.Get(ctx, PresenceDescriptionHistoryKey)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var history []string
	if err := json.Unmarshal([]byte(raw), &history); err != nil {
		log.Printf("presence change: corrupt history json, ignoring: %v", err)
		return nil, nil
	}
	if len(history) > maxPresenceDescriptionHistory {
		history = history[:maxPresenceDescriptionHistory]
	}
	return history, nil
}

func (uc *PresenceChangeUseCase) recordHistory(ctx context.Context, description string) error {
	uc.historyMu.Lock()
	defer uc.historyMu.Unlock()
	if uc.settings == nil {
		return nil
	}
	history, err := uc.loadHistory(ctx)
	if err != nil {
		return err
	}
	history = prependUniqueHistory(history, description)
	b, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return uc.settings.Set(ctx, PresenceDescriptionHistoryKey, string(b))
}

func prependUniqueHistory(history []string, item string) []string {
	out := make([]string, 0, len(history)+1)
	seen := make(map[string]struct{}, len(history)+1)
	add := func(entry string) {
		if _, ok := seen[entry]; ok {
			return
		}
		seen[entry] = struct{}{}
		out = append(out, entry)
	}
	add(item)
	for _, h := range history {
		add(h)
		if len(out) >= maxPresenceDescriptionHistory {
			break
		}
	}
	return out
}
