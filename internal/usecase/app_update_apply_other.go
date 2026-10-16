//go:build !windows

package usecase

import (
	"context"
	"errors"
)

// ApplyUpdate is only supported on Windows.
func (uc *AppUpdateUseCase) ApplyUpdate(_ context.Context, _ string) error {
	return errors.New(ErrAppUpdateUnsupportedSent)
}
