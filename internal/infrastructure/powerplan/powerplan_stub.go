//go:build !windows

package powerplan

import "fmt"

// Plan is an OS power scheme.
type Plan struct {
	GUID string
	Name string
}

// ListDetected is unsupported off Windows.
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
