package usecase

import "context"

// ponytail:#129 domain AppSettingsRepository removed; boundary stays usecase-local.
type appSettingsRepo interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
}
