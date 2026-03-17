package settings

import "context"

// AppSettingsRepository defines persistence for app settings (key-value).
type AppSettingsRepository interface {
	// Get returns the value for the given key, or empty string if not found.
	Get(ctx context.Context, key string) (string, error)
	// Set saves a key-value pair.
	Set(ctx context.Context, key, value string) error
	// GetAll returns all settings.
	GetAll(ctx context.Context) (map[string]string, error)
}
