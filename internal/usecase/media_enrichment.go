package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

// TryEnrichAfterIngest attempts enrichment for a newly ingested screenshot when auto mode is on.
func (uc *MediaUseCase) TryEnrichAfterIngest(ctx context.Context, screenshotID string) {
	if uc.enrichment == nil || !uc.galleryAutoEnrichEnabled(ctx) {
		return
	}
	if _, err := uc.EnrichScreenshot(ctx, screenshotID, true); err != nil {
		log.Printf("[enrich] ingest enrich failed for %s: %v", screenshotID, err)
	}
}

// RetryPendingEnrichments re-attempts pending and no_match enrichments.
func (uc *MediaUseCase) RetryPendingEnrichments(ctx context.Context) (int, error) {
	if uc.enrichment == nil {
		return 0, nil
	}
	count := 0
	for _, status := range []string{media.EnrichmentStatusPending, media.EnrichmentStatusNoMatch} {
		if err := ctx.Err(); err != nil {
			return count, err
		}
		rows, err := uc.enrichment.ListByStatus(ctx, status)
		if err != nil {
			return count, err
		}
		for _, row := range rows {
			if err := ctx.Err(); err != nil {
				return count, err
			}
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
	return uc.enrichScreenshotRow(ctx, s, writeFile, nil)
}

type screenshotFileMeta struct {
	fields      media.EnrichmentFields
	hasMetaDate bool
}

func readScreenshotFileMeta(path string) (screenshotFileMeta, error) {
	fields, err := media.ReadEnrichmentFieldsFromFile(path)
	if err != nil {
		return screenshotFileMeta{}, err
	}
	hasMetaDate, err := media.HasMetadataTakenAt(path)
	if err != nil {
		return screenshotFileMeta{}, err
	}
	return screenshotFileMeta{fields: fields, hasMetaDate: hasMetaDate}, nil
}

func (uc *MediaUseCase) enrichScreenshotRow(ctx context.Context, s *media.Screenshot, writeFile bool, fileMeta *screenshotFileMeta) (*EnrichScreenshotResult, error) {
	existing, err := uc.enrichment.GetByScreenshotID(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	var meta screenshotFileMeta
	if fileMeta != nil {
		meta = *fileMeta
	} else {
		meta, err = readScreenshotFileMeta(s.FilePath)
		if err != nil {
			return nil, err
		}
	}
	if !media.IsEligibleForEnrichment(s.WorldID, meta.hasMetaDate, meta.fields, existing) {
		return nil, nil
	}
	if s.TakenAt == nil {
		row := uc.pendingEnrichment(s.ID, "missing_taken_at")
		if err := uc.enrichment.Save(ctx, row); err != nil {
			return nil, err
		}
		return enrichResultFromRow(row), nil
	}

	extracted, extractErr := extractScreenshotMetadata(s.FilePath)
	if extractErr != nil {
		log.Printf("[enrich] metadata extract failed for %s: %v", s.FilePath, extractErr)
	}
	outcome := uc.correlateForScreenshot(ctx, s, extracted.WorldID, meta.fields.InstanceID, *s.TakenAt)
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
			row.Status = media.EnrichmentStatusConflict
			row.SkipReason = "metadata_write_failed"
		} else {
			now := time.Now()
			row.EnrichedAt = &now
		}
	} else if outcome.Status == media.EnrichmentStatusSuccess {
		now := time.Now()
		row.EnrichedAt = &now
	}
	if err := uc.enrichment.Save(ctx, row); err != nil {
		return nil, err
	}
	return enrichResultFromRow(row), nil
}

func (uc *MediaUseCase) correlateForScreenshot(ctx context.Context, s *media.Screenshot, xmpWorldID, existingInstanceID string, takenAt time.Time) media.CorrelationOutcome {
	sessions, err := uc.playSessions.ListOverlappingAt(ctx, takenAt)
	if err != nil {
		return media.CorrelationOutcome{Status: media.EnrichmentStatusPending, SkipReason: "activity_query_failed"}
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
		return media.CorrelationOutcome{Status: media.EnrichmentStatusPending, SkipReason: "encounter_query_failed"}
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
	return strings.TrimSpace(id)
}

// GetScreenshotEnrichment returns persisted enrichment for UI.
func (uc *MediaUseCase) GetScreenshotEnrichment(ctx context.Context, screenshotID string) (*media.ScreenshotEnrichment, error) {
	if uc.enrichment == nil {
		return nil, nil
	}
	return uc.enrichment.GetByScreenshotID(ctx, trimID(screenshotID))
}
