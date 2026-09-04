package media

import "time"

// Enrichment status values persisted in screenshot_enrichment.status.
const (
	EnrichmentStatusPending   = "pending"
	EnrichmentStatusSuccess   = "success"
	EnrichmentStatusNoMatch   = "no_match"
	EnrichmentStatusConflict  = "conflict"
	EnrichmentStatusAmbiguous = "ambiguous"
)

// EnrichmentParticipant is one user present in an instance at screenshot time.
type EnrichmentParticipant struct {
	VRCUserID   string `json:"vrcUserId"`
	DisplayName string `json:"displayName"`
}

// EnrichmentFields are vrctweaker-oriented metadata read from or written to XMP.
type EnrichmentFields struct {
	InstanceID   string
	Participants []EnrichmentParticipant
}

// ScreenshotEnrichment is the persisted enrichment outcome for a screenshot.
type ScreenshotEnrichment struct {
	ScreenshotID     string
	Status           string
	InstanceID       string
	WorldID          string
	ParticipantsJSON string
	SkipReason       string
	EnrichedAt       *time.Time
}

// SessionOverlap holds a play session that overlaps a point in time.
type SessionOverlap struct {
	InstanceID string
	WorldID    string
}
