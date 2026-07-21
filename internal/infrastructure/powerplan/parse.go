package powerplan

import (
	"fmt"
	"regexp"
	"strings"
)

// Well-known Windows scheme GUIDs (locale-independent).
var presetGUIDs = map[string]string{
	"power_saver":      "a1841308-3541-4fab-bc81-f71556f20b4a",
	"balanced":         "381b4222-f694-41f0-9685-ff5bb260df2e",
	"high_performance": "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c",
}

// GUID pattern used by powercfg (locale-independent).
var schemeGUIDRE = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

var exactGUIDRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidGUID reports whether s is a bare Windows power-scheme GUID.
func ValidGUID(s string) bool {
	return exactGUIDRE.MatchString(strings.TrimSpace(s))
}

func parseListOutput(s string) ([]Plan, error) {
	var plans []Plan
	seen := map[string]struct{}{}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		loc := schemeGUIDRE.FindStringIndex(line)
		if loc == nil {
			continue
		}
		guid := line[loc[0]:loc[1]]
		key := strings.ToLower(guid)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		name := extractPlanName(line[loc[1]:])
		plans = append(plans, Plan{GUID: guid, Name: name})
	}
	if len(plans) == 0 {
		return nil, fmt.Errorf("no power plans found")
	}
	return plans, nil
}

func extractPlanName(rest string) string {
	rest = strings.TrimSpace(rest)
	if i := strings.Index(rest, "("); i >= 0 {
		if j := strings.Index(rest[i+1:], ")"); j >= 0 {
			return strings.TrimSpace(rest[i+1 : i+1+j])
		}
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	name := strings.Join(fields, " ")
	name = strings.TrimSuffix(name, "*")
	return strings.TrimSpace(name)
}

func resolvePresetFromPlans(preset string, plans []Plan) (string, error) {
	guid, ok := presetGUIDs[preset]
	if !ok {
		return "", fmt.Errorf("unknown preset %q", preset)
	}
	for _, p := range plans {
		if strings.EqualFold(p.GUID, guid) {
			return p.GUID, nil
		}
	}
	return "", fmt.Errorf("preset %q not found on system", preset)
}
