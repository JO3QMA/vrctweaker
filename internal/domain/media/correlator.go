package media

import (
	"sort"
	"time"
)

// CorrelationOutcome is the result of matching a screenshot time to activity.
type CorrelationOutcome struct {
	Status       string
	InstanceID   string
	WorldID      string
	SkipReason   string
	Participants []EnrichmentParticipant
}

// SessionAtTime describes one play session candidate for correlation.
type SessionAtTime struct {
	InstanceID string
	WorldID    string
	StartTime  time.Time
	EndTime    *time.Time
}

// EncounterAtTime describes one user encounter for participant snapshots.
type EncounterAtTime struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	JoinedAt    time.Time
	LeftAt      *time.Time
}

// SessionOverlapsAt reports whether a session contains t (inclusive bounds).
func SessionOverlapsAt(s SessionAtTime, t time.Time) bool {
	if t.Before(s.StartTime) {
		return false
	}
	if s.EndTime != nil && t.After(*s.EndTime) {
		return false
	}
	return s.InstanceID != ""
}

// CorrelateActivity picks the play session and participant snapshot for a screenshot.
func CorrelateActivity(
	takenAt time.Time,
	filterWorldID string,
	xmpWorldID string,
	existingInstanceID string,
	sessions []SessionAtTime,
	encounters []EncounterAtTime,
) CorrelationOutcome {
	if xmpWorldID != "" && filterWorldID != "" && xmpWorldID != filterWorldID {
		return CorrelationOutcome{
			Status:     EnrichmentStatusConflict,
			SkipReason: "xmp_world_id_mismatch",
		}
	}
	worldFilter := filterWorldID
	if worldFilter == "" {
		worldFilter = xmpWorldID
	}

	var candidates []SessionAtTime
	for _, s := range sessions {
		if !SessionOverlapsAt(s, takenAt) {
			continue
		}
		if worldFilter != "" && s.WorldID != "" && s.WorldID != worldFilter {
			continue
		}
		candidates = append(candidates, s)
	}
	switch len(candidates) {
	case 0:
		return CorrelationOutcome{Status: EnrichmentStatusNoMatch, SkipReason: "no_play_session"}
	case 1:
		chosen := candidates[0]
		if existingInstanceID != "" && existingInstanceID != chosen.InstanceID {
			return CorrelationOutcome{
				Status:     EnrichmentStatusConflict,
				SkipReason: "instance_id_mismatch",
			}
		}
		return CorrelationOutcome{
			Status:       EnrichmentStatusSuccess,
			InstanceID:   chosen.InstanceID,
			WorldID:      chosen.WorldID,
			Participants: SnapshotParticipants(encounters, chosen.InstanceID, takenAt),
		}
	default:
		return CorrelationOutcome{
			Status:     EnrichmentStatusAmbiguous,
			SkipReason: "multiple_play_sessions",
		}
	}
}

// SnapshotParticipants returns users in instance at time t (joined, not yet left).
func SnapshotParticipants(encounters []EncounterAtTime, instanceID string, t time.Time) []EnrichmentParticipant {
	latest := make(map[string]EnrichmentParticipant)
	latestJoin := make(map[string]time.Time)
	for _, e := range encounters {
		if e.InstanceID != instanceID {
			continue
		}
		if t.Before(e.JoinedAt) {
			continue
		}
		if e.LeftAt != nil && !t.Before(*e.LeftAt) {
			continue
		}
		prev, ok := latestJoin[e.VRCUserID]
		if !ok || e.JoinedAt.After(prev) {
			latestJoin[e.VRCUserID] = e.JoinedAt
			latest[e.VRCUserID] = EnrichmentParticipant{
				VRCUserID:   e.VRCUserID,
				DisplayName: e.DisplayName,
			}
		}
	}
	out := make([]EnrichmentParticipant, 0, len(latest))
	for _, p := range latest {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].VRCUserID < out[j].VRCUserID
	})
	return out
}

// IsEligibleForEnrichment reports whether a screenshot should be considered for enrichment.
func IsEligibleForEnrichment(
	worldID string,
	hasMetadataTakenAt bool,
	fileFields EnrichmentFields,
	enrichment *ScreenshotEnrichment,
) bool {
	if enrichment != nil && enrichment.Status == EnrichmentStatusSuccess {
		return false
	}
	if worldID == "" {
		return true
	}
	if !hasMetadataTakenAt {
		return true
	}
	if fileFields.InstanceID == "" {
		if enrichment == nil || enrichment.Status != EnrichmentStatusSuccess {
			return true
		}
	}
	return false
}
