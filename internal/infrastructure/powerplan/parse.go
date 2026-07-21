package powerplan

import (
	"fmt"
	"strings"
)

// Well-known Windows scheme GUIDs (locale-independent).
var presetGUIDs = map[string]string{
	"power_saver":      "a1841308-3541-4fab-bc81-f71556f20b4a",
	"balanced":         "381b4222-f694-41f0-9685-ff5bb260df2e",
	"high_performance": "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c",
}

func parseListOutput(s string) ([]Plan, error) {
	var plans []Plan
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "Power Scheme GUID:") {
			continue
		}
		i := strings.Index(line, ":")
		if i < 0 {
			continue
		}
		rest := strings.TrimSpace(line[i+1:])
		parts := strings.SplitN(rest, "  ", 2)
		guid := strings.TrimSpace(parts[0])
		name := ""
		if len(parts) > 1 {
			name = strings.TrimSpace(parts[1])
			name = strings.TrimSuffix(name, " *")
			name = strings.TrimSpace(name)
			name = strings.Trim(name, "()")
		}
		if guid != "" {
			plans = append(plans, Plan{GUID: guid, Name: name})
		}
	}
	if len(plans) == 0 {
		return nil, fmt.Errorf("no power plans found")
	}
	return plans, nil
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
