package usecase

import "vrchat-tweaker/internal/domain/launcher"

// ResolveInstanceRejoinProfileID picks the initial Instance rejoin launch profile:
// Last launch profile → Default launch profile → first item in profiles (list order).
func ResolveInstanceRejoinProfileID(profiles []*launcher.LaunchProfile, lastID string) string {
	if len(profiles) == 0 {
		return ""
	}
	if lastID != "" {
		for _, p := range profiles {
			if p.ID == lastID {
				return lastID
			}
		}
	}
	for _, p := range profiles {
		if p.IsDefault {
			return p.ID
		}
	}
	return profiles[0].ID
}
