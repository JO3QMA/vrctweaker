package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"vrchat-tweaker/internal/domain/media"
)

// ScreenshotEnrichmentRepository persists screenshot enrichment rows.
type ScreenshotEnrichmentRepository struct {
	db *sql.DB
}

// NewScreenshotEnrichmentRepository creates a ScreenshotEnrichmentRepository.
func NewScreenshotEnrichmentRepository(db *sql.DB) *ScreenshotEnrichmentRepository {
	return &ScreenshotEnrichmentRepository{db: db}
}

// GetByScreenshotID returns enrichment for a screenshot or nil.
func (r *ScreenshotEnrichmentRepository) GetByScreenshotID(ctx context.Context, screenshotID string) (*media.ScreenshotEnrichment, error) {
	row := r.db.QueryRowContext(ctx, `SELECT screenshot_id, status, IFNULL(instance_id,''), IFNULL(world_id,''),
		IFNULL(participants_json,''), IFNULL(skip_reason,''), enriched_at
		FROM screenshot_enrichment WHERE screenshot_id = ?`, screenshotID)
	return scanEnrichmentRow(row)
}

// Save upserts enrichment for a screenshot.
func (r *ScreenshotEnrichmentRepository) Save(ctx context.Context, e *media.ScreenshotEnrichment) error {
	if e == nil {
		return nil
	}
	enrichedAt := nullableTime(e.EnrichedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO screenshot_enrichment
		(screenshot_id, status, instance_id, world_id, participants_json, skip_reason, enriched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(screenshot_id) DO UPDATE SET
		status = excluded.status, instance_id = excluded.instance_id, world_id = excluded.world_id,
		participants_json = excluded.participants_json, skip_reason = excluded.skip_reason,
		enriched_at = excluded.enriched_at`,
		e.ScreenshotID, e.Status, nullString(e.InstanceID), nullString(e.WorldID),
		nullString(e.ParticipantsJSON), nullString(e.SkipReason), enrichedAt)
	return err
}

// ListByStatus returns enrichments with the given status.
func (r *ScreenshotEnrichmentRepository) ListByStatus(ctx context.Context, status string) ([]*media.ScreenshotEnrichment, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT screenshot_id, status, IFNULL(instance_id,''), IFNULL(world_id,''),
		IFNULL(participants_json,''), IFNULL(skip_reason,''), enriched_at
		FROM screenshot_enrichment WHERE status = ?`, status)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []*media.ScreenshotEnrichment
	for rows.Next() {
		e, err := scanEnrichmentRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ParticipantsFromJSON decodes participants_json for tests and callers.
func ParticipantsFromJSON(raw string) ([]media.EnrichmentParticipant, error) {
	if raw == "" {
		return nil, nil
	}
	var out []media.EnrichmentParticipant
	err := json.Unmarshal([]byte(raw), &out)
	return out, err
}

func scanEnrichmentRow(row *sql.Row) (*media.ScreenshotEnrichment, error) {
	var id, status, instanceID, worldID, participantsJSON, skipReason string
	var enrichedAt sql.NullString
	err := row.Scan(&id, &status, &instanceID, &worldID, &participantsJSON, &skipReason, &enrichedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &media.ScreenshotEnrichment{
		ScreenshotID:     id,
		Status:           status,
		InstanceID:       instanceID,
		WorldID:          worldID,
		ParticipantsJSON: participantsJSON,
		SkipReason:       skipReason,
		EnrichedAt:       parseTime(enrichedAt),
	}, nil
}

func scanEnrichmentRows(rows *sql.Rows) (*media.ScreenshotEnrichment, error) {
	var id, status, instanceID, worldID, participantsJSON, skipReason string
	var enrichedAt sql.NullString
	if err := rows.Scan(&id, &status, &instanceID, &worldID, &participantsJSON, &skipReason, &enrichedAt); err != nil {
		return nil, err
	}
	return &media.ScreenshotEnrichment{
		ScreenshotID:     id,
		Status:           status,
		InstanceID:       instanceID,
		WorldID:          worldID,
		ParticipantsJSON: participantsJSON,
		SkipReason:       skipReason,
		EnrichedAt:       parseTime(enrichedAt),
	}, nil
}
