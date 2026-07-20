//go:build windows

package powerplan

import (
	"fmt"
	"os/exec"
	"strings"
)

// ListDetected returns power schemes from powercfg /list.
func ListDetected() ([]Plan, error) {
	out, err := exec.Command("powercfg", "/list").Output()
	if err != nil {
		return nil, err
	}
	return parseListOutput(string(out))
}

// SetActive activates a scheme by GUID.
func SetActive(guid string) error {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return fmt.Errorf("empty power plan guid")
	}
	_, err := exec.Command("powercfg", "/setactive", guid).CombinedOutput()
	return err
}

// ResolvePreset maps a preset key to a GUID on this machine.
func ResolvePreset(preset string) (string, error) {
	plans, err := ListDetected()
	if err != nil {
		return "", err
	}
	want := presetAliases[preset]
	if want == "" {
		return "", fmt.Errorf("unknown preset %q", preset)
	}
	for _, p := range plans {
		if strings.EqualFold(p.Name, want) {
			return p.GUID, nil
		}
	}
	// Substring match for localized names.
	for _, p := range plans {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(want)) {
			return p.GUID, nil
		}
	}
	return "", fmt.Errorf("preset %q not found on system", preset)
}

var presetAliases = map[string]string{
	"power_saver":      "Power saver",
	"balanced":         "Balanced",
	"high_performance": "High performance",
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
			name = strings.Trim(strings.TrimSpace(parts[1]), "()")
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
