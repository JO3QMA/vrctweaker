//go:build windows

package powerplan

import (
	"fmt"
	"os/exec"
	"strings"
)

// ListDetected returns power schemes from powercfg /list.
func ListDetected() ([]Plan, error) {
	out, err := exec.Command("powercfg", "/list").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("powercfg /list failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return parseListOutput(string(out))
}

// SetActive activates a scheme by GUID.
func SetActive(guid string) error {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return fmt.Errorf("empty power plan guid")
	}
	out, err := exec.Command("powercfg", "/setactive", guid).CombinedOutput()
	if err != nil {
		return fmt.Errorf("powercfg /setactive %s failed: %w (output: %s)", guid, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ResolvePreset maps a preset key to a GUID on this machine.
func ResolvePreset(preset string) (string, error) {
	plans, err := ListDetected()
	if err != nil {
		return "", err
	}
	return resolvePresetFromPlans(preset, plans)
}
