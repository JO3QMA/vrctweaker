//go:build !windows

package powerplan

import (
	"context"
	"fmt"
)

// ListDetected returns an empty list off Windows (not an error; ADR: unsupported is normal).
func ListDetected(context.Context) ([]Plan, error) {
	return nil, nil
}

// SetActive is unsupported off Windows.
func SetActive(context.Context, string) error {
	return fmt.Errorf("set_power_plan: unsupported platform")
}

// ResolvePreset is unsupported off Windows.
func ResolvePreset(context.Context, string) (string, error) {
	return "", fmt.Errorf("set_power_plan: unsupported platform")
}
