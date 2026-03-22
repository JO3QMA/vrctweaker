package activity

import "strings"

// WorldIDFromInstanceKey returns the wrld_* prefix from an instance key like "wrld_uuid:123~...".
func WorldIDFromInstanceKey(instanceKey string) string {
	instanceKey = strings.TrimSpace(instanceKey)
	if !strings.HasPrefix(instanceKey, "wrld_") {
		return ""
	}
	i := strings.Index(instanceKey, ":")
	if i <= 0 {
		return ""
	}
	return instanceKey[:i]
}
