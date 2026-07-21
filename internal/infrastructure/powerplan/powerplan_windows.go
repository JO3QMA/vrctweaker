//go:build windows

package powerplan

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const powercfgTimeout = 15 * time.Second

func withPowercfgTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, powercfgTimeout)
}

// ListDetected returns power schemes from powercfg /list.
func ListDetected(ctx context.Context) ([]Plan, error) {
	ctx, cancel := withPowercfgTimeout(ctx)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powercfg", "/list")
	hideConsoleWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("powercfg /list failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return parseListOutput(string(out))
}

// SetActive activates a scheme by GUID.
func SetActive(ctx context.Context, guid string) error {
	guid = strings.TrimSpace(guid)
	if !ValidGUID(guid) {
		return fmt.Errorf("invalid power plan guid")
	}
	ctx, cancel := withPowercfgTimeout(ctx)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powercfg", "/setactive", guid)
	hideConsoleWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powercfg /setactive %s failed: %w (output: %s)", guid, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ResolvePreset maps a preset key to a GUID on this machine.
func ResolvePreset(ctx context.Context, preset string) (string, error) {
	plans, err := ListDetected(ctx)
	if err != nil {
		return "", err
	}
	return resolvePresetFromPlans(preset, plans)
}
