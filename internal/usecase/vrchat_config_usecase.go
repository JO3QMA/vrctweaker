package usecase

import (
	"vrchat-tweaker/internal/domain/vrchatconfig"
)

// VRChatConfigUseCase handles VRChat config.json operations.
type VRChatConfigUseCase struct {
	repo vrchatconfig.ConfigRepository
}

// NewVRChatConfigUseCase creates a new VRChatConfigUseCase.
func NewVRChatConfigUseCase(repo vrchatconfig.ConfigRepository) *VRChatConfigUseCase {
	return &VRChatConfigUseCase{repo: repo}
}

// Exists checks if config.json exists.
func (uc *VRChatConfigUseCase) Exists() (bool, error) {
	return uc.repo.Exists()
}

// Get reads the current config.json.
func (uc *VRChatConfigUseCase) Get() (*vrchatconfig.VRChatConfig, error) {
	return uc.repo.Read()
}

// Save writes config.json.
func (uc *VRChatConfigUseCase) Save(cfg *vrchatconfig.VRChatConfig) error {
	return uc.repo.Write(cfg)
}

// Delete removes config.json.
func (uc *VRChatConfigUseCase) Delete() error {
	return uc.repo.Delete()
}
