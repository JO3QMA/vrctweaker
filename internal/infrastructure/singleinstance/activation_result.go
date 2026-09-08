package singleinstance

import "fmt"

// combineNotifyExistingResults returns nil when either the cross-process signal or
// native window activation succeeds. Both paths are attempted so foreground focus
// works even when SetEvent succeeds but the running instance cannot steal focus alone.
func combineNotifyExistingResults(signalErr, nativeErr error) error {
	if signalErr == nil || nativeErr == nil {
		return nil
	}
	return fmt.Errorf("singleinstance: existing instance could not be activated (signal: %w; native: %w)", signalErr, nativeErr)
}
