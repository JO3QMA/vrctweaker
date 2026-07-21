//go:build !windows

package powerplan

import "fmt"

// ListDetected returns an empty list off Windows (not an error; ADR: unsupported is normal).
func ListDetected() ([]Plan, error) {
	return nil, nil
}

// SetActive is unsupported off Windows.
func SetActive(string) error {
	return fmt.Errorf("set_power_plan: unsupported platform")
}

// ResolvePreset is unsupported off Windows.
func ResolvePreset(string) (string, error) {
	return "", fmt.Errorf("set_power_plan: unsupported platform")
}
