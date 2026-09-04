package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/media"
)

const keyGalleryAutoEnrichMetadata = "gallery_auto_enrich_metadata"

// EnrichScreenshotResult is the outcome of enriching one screenshot.
type EnrichScreenshotResult struct {
	ScreenshotID string
	Status       string
	SkipReason   string
	InstanceID   string
}

// EnrichBatchResult aggregates batch enrichment outcomes.
type EnrichBatchResult struct {
	Processed int
	Results   []EnrichScreenshotResult
}

// MediaEnrichmentDeps wires activity and enrichment persistence into MediaUseCase.
type MediaEnrichmentDeps struct {
	PlaySessions playSessionRepo
	Encounters   userEncounterRepo
	Enrichment   screenshotEnrichmentRepo
	Settings     appSettingsRepo
}

// SetEnrichmentDeps attaches optional enrichment collaborators.
func (uc *MediaUseCase) SetEnrichmentDeps(deps MediaEnrichmentDeps) {
	uc.playSessions = deps.PlaySessions
	uc.encounters = deps.Encounters
	uc.enrichment = deps.Enrichment
	uc.enrichSettings = deps.Settings
}

func (uc *MediaUseCase) galleryAutoEnrichEnabled(ctx context.Context) bool {
	if uc.enrichSettings == nil {
		return true
	}
	v, err := uc.enrichSettings.Get(ctx, keyGalleryAutoEnrichMetadata)
	if err != nil || v == "" {
		return true
	}
	return v == "1" || v == "true"
}

// QueueEnrichmentAfterIngest attempts enrichment for a newly ingested screenshot when auto mode is on.
func (uc *MediaUseCase) QueueEnrichmentAfterIngest(ctx context.Context, screenshotID string) {
	if uc.enrichment == nil || !uc.galleryAutoEnrichEnabled(ctx) {
		return
	}
	_, _ = uc.EnrichScreenshot(ctx, screenshotID, true)
}

// RetryPendingEnrichments re-attempts pending and no_match enrichments.
func (uc *MediaUseCase) RetryPendingEnrichments(ctx context.Context) (int, error) {
	if uc.enrichment == nil {
		return 0, nil
	}
	count := 0
	for _, status := range []string{media.EnrichmentStatusPending, media.EnrichmentStatusNoMatch} {
		rows, err := uc.enrichment.ListByStatus(ctx, status)
		if err != nil {
			return count, err
		}
		for _, row := range rows {
			res, err := uc.EnrichScreenshot(ctx, row.ScreenshotID, false)
			if err != nil {
				return count, err
			}
			if res != nil {
				count++
			}
		}
	}
	return count, nil
}

// EnrichScreenshot correlates activity and writes instance/participant metadata.
func (uc *MediaUseCase) EnrichScreenshot(ctx context.Context, screenshotID string, writeFile bool) (*EnrichScreenshotResult, error) {
	if uc.enrichment == nil || uc.playSessions == nil || uc.encounters == nil {
		return nil, fmt.Errorf("enrichment not configured")
	}
	screenshotID = trimID(screenshotID)
	if screenshotID == "" {
		return nil, fmt.Errorf("screenshot id required")
	}
	s, err := uc.repo.GetByID(ctx, screenshotID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, errScreenshotNotFound
	}
	return uc.enrichScreenshotRow(ctx, s, writeFile)
}

// EnrichEligibleScreenshots enriches eligible screenshots (all when ids empty).
func (uc *MediaUseCase) EnrichEligibleScreenshots(ctx context.Context, ids []string, writeFile bool) (*EnrichBatchResult, error) {
	if uc.enrichment == nil {
		return nil, fmt.Errorf("enrichment not configured")
	}
	var targets []*media.Screenshot
	if len(ids) > 0 {
		for _, id := range ids {
			s, err := uc.repo.GetByID(ctx, trimID(id))
			if err != nil {
				return nil, err
			}
			if s != nil {
				targets = append(targets, s)
			}
		}
	} else {
		list, err := uc.repo.List(ctx, nil)
		if err != nil {
			return nil, err
		}
		targets = list
	}
	out := &EnrichBatchResult{}
	for _, s := range targets {
		eligible, err := uc.isScreenshotEligible(ctx, s)
		if err != nil {
			return out, err
		}
		if !eligible {
			continue
		}
		res, err := uc.enrichScreenshotRow(ctx, s, writeFile)
		if err != nil {
			return out, err
		}
		if res != nil {
			out.Processed++
			out.Results = append(out.Results, *res)
		}
	}
	return out, nil
}

func (uc *MediaUseCase) enrichScreenshotRow(ctx context.Context, s *media.Screenshot, writeFile bool) (*EnrichScreenshotResult, error) {
	existing, _ := uc.enrichment.GetByScreenshotID(ctx, s.ID)
	fileFields, err := media.ReadEnrichmentFieldsFromFile(s.FilePath)
	if err != nil {
		return nil, err
	}
	hasMetaDate, err := media.HasMetadataTakenAt(s.FilePath)
	if err != nil {
		return nil, err
	}
	if !media.IsEligibleForEnrichment(s.WorldID, hasMetaDate, fileFields, existing) {
		return nil, nil
	}
	if s.TakenAt == nil {
		row := uc.pendingEnrichment(s.ID, "missing_taken_at")
		_ = uc.enrichment.Save(ctx, row)
		return enrichResultFromRow(row), nil
	}

	meta, _ := extractScreenshotMetadata(s.FilePath)
	outcome := uc.correlateForScreenshot(ctx, s, meta.WorldID, fileFields.InstanceID, *s.TakenAt)
	row := &media.ScreenshotEnrichment{
		ScreenshotID: s.ID,
		Status:       outcome.Status,
		InstanceID:   outcome.InstanceID,
		WorldID:      outcome.WorldID,
		SkipReason:   outcome.SkipReason,
	}
	if len(outcome.Participants) > 0 {
		raw, err := json.Marshal(outcome.Participants)
		if err != nil {
			return nil, err
		}
		row.ParticipantsJSON = string(raw)
	}
	if outcome.Status == media.EnrichmentStatusSuccess && writeFile {
		fields := media.EnrichmentFields{
			InstanceID:   outcome.InstanceID,
			Participants: outcome.Participants,
		}
		if err := media.EmbedEnrichmentMetadata(s.FilePath, fields); err != nil {
			return nil, err
		}
		now := time.Now()
		row.EnrichedAt = &now
	} else if outcome.Status == media.EnrichmentStatusSuccess {
		now := time.Now()
		row.EnrichedAt = &now
	} else if outcome.Status == media.EnrichmentStatusNoMatch || outcome.Status == media.EnrichmentStatusAmbiguous || outcome.Status == media.EnrichmentStatusConflict {
		// keep enriched_at nil
	} else if outcome.Status == "" {
		row.Status = media.EnrichmentStatusPending
	}
	if err := uc.enrichment.Save(ctx, row); err != nil {
		return nil, err
	}
	return enrichResultFromRow(row), nil
}

func (uc *MediaUseCase) correlateForScreenshot(ctx context.Context, s *media.Screenshot, xmpWorldID, existingInstanceID string, takenAt time.Time) media.CorrelationOutcome {
	sessions, err := uc.playSessions.ListOverlappingAt(ctx, takenAt)
	if err != nil {
		return media.CorrelationOutcome{Status: media.EnrichmentStatusNoMatch, SkipReason: "activity_query_failed"}
	}
	sessionDTOs := make([]media.SessionAtTime, 0, len(sessions))
	for _, ps := range sessions {
		sessionDTOs = append(sessionDTOs, media.SessionAtTime{
			InstanceID: ps.InstanceID,
			WorldID:    activity.WorldIDFromInstanceKey(ps.InstanceID),
			StartTime:  ps.StartTime,
			EndTime:    ps.EndTime,
		})
	}
	from := takenAt.Add(-24 * time.Hour)
	to := takenAt.Add(24 * time.Hour)
	encounters, err := uc.encounters.List(ctx, &activity.EncounterFilter{From: &from, To: &to})
	if err != nil {
		return media.CorrelationOutcome{Status: media.EnrichmentStatusNoMatch, SkipReason: "encounter_query_failed"}
	}
	encDTOs := make([]media.EncounterAtTime, 0, len(encounters))
	for _, e := range encounters {
		encDTOs = append(encDTOs, media.EncounterAtTime{
			VRCUserID:   e.VRCUserID,
			DisplayName: e.DisplayName,
			InstanceID:  e.InstanceID,
			JoinedAt:    e.JoinedAt,
			LeftAt:      e.LeftAt,
		})
	}
	filterWorld := s.WorldID
	if filterWorld == "" {
		filterWorld = xmpWorldID
	}
	return media.CorrelateActivity(takenAt, filterWorld, xmpWorldID, existingInstanceID, sessionDTOs, encDTOs)
}

func (uc *MediaUseCase) isScreenshotEligible(ctx context.Context, s *media.Screenshot) (bool, error) {
	existing, err := uc.enrichment.GetByScreenshotID(ctx, s.ID)
	if err != nil {
		return false, err
	}
	fileFields, err := media.ReadEnrichmentFieldsFromFile(s.FilePath)
	if err != nil {
		return false, err
	}
	hasMetaDate, err := media.HasMetadataTakenAt(s.FilePath)
	if err != nil {
		return false, err
	}
	return media.IsEligibleForEnrichment(s.WorldID, hasMetaDate, fileFields, existing), nil
}

func (uc *MediaUseCase) pendingEnrichment(screenshotID, reason string) *media.ScreenshotEnrichment {
	return &media.ScreenshotEnrichment{
		ScreenshotID: screenshotID,
		Status:       media.EnrichmentStatusPending,
		SkipReason:   reason,
	}
}

func enrichResultFromRow(row *media.ScreenshotEnrichment) *EnrichScreenshotResult {
	if row == nil {
		return nil
	}
	return &EnrichScreenshotResult{
		ScreenshotID: row.ScreenshotID,
		Status:       row.Status,
		SkipReason:   row.SkipReason,
		InstanceID:   row.InstanceID,
	}
}

func trimID(id string) string {
	for len(id) > 0 && (id[0] == ' ' || id[0] == '\t') {
		id = id[1:]
	}
	for len(id) > 0 && (id[len(id)-1] == ' ' || id[len(id)-1] == '\t') {
		id = id[:len(id)-1]
	}
	return id
}

// GetScreenshotEnrichment returns persisted enrichment for UI.
func (uc *MediaUseCase) GetScreenshotEnrichment(ctx context.Context, screenshotID string) (*media.ScreenshotEnrichment, error) {
	if uc.enrichment == nil {
		return nil, nil
	}
	return uc.enrichment.GetByScreenshotID(ctx, trimID(screenshotID))
}
