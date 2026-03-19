package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

// VRChatConfigFileRepository reads/writes VRChat config.json from the filesystem.
type VRChatConfigFileRepository struct {
	path string
}

// NewVRChatConfigFileRepository creates a repository for the given config.json path.
func NewVRChatConfigFileRepository(path string) *VRChatConfigFileRepository {
	return &VRChatConfigFileRepository{path: path}
}

func (r *VRChatConfigFileRepository) Exists() (bool, error) {
	info, err := os.Stat(r.path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.Mode().IsRegular(), nil
}

func (r *VRChatConfigFileRepository) Read() (*vrchatconfig.VRChatConfig, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config.json does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}
	var cfg vrchatconfig.VRChatConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}
	return &cfg, nil
}

func (r *VRChatConfigFileRepository) Write(cfg *vrchatconfig.VRChatConfig) error {
	if cfg == nil {
		cfg = &vrchatconfig.VRChatConfig{}
	}
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(r.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config.json: %w", err)
	}
	return nil
}

func (r *VRChatConfigFileRepository) Delete() error {
	if err := os.Remove(r.path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete config.json: %w", err)
	}
	return nil
}
