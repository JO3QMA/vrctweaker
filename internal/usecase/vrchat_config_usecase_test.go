package usecase

import (
	"fmt"
	"testing"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

type fakeConfigRepo struct {
	cfg     *vrchatconfig.VRChatConfig
	deleted bool
}

func (f *fakeConfigRepo) Exists() (bool, error) {
	return f.cfg != nil && !f.deleted, nil
}

func (f *fakeConfigRepo) Read() (*vrchatconfig.VRChatConfig, error) {
	if f.cfg == nil || f.deleted {
		return nil, fmt.Errorf("config.json does not exist")
	}
	return f.cfg, nil
}

func (f *fakeConfigRepo) Write(cfg *vrchatconfig.VRChatConfig) error {
	f.cfg = cfg
	f.deleted = false
	return nil
}

func (f *fakeConfigRepo) Delete() error {
	f.cfg = nil
	f.deleted = true
	return nil
}

func TestVRChatConfigUseCase_Exists_False(t *testing.T) {
	repo := &fakeConfigRepo{}
	uc := NewVRChatConfigUseCase(repo)

	exists, err := uc.Exists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false")
	}
}

func TestVRChatConfigUseCase_SaveAndGet(t *testing.T) {
	repo := &fakeConfigRepo{}
	uc := NewVRChatConfigUseCase(repo)

	cfg := &vrchatconfig.VRChatConfig{
		CameraResWidth:  2560,
		CameraResHeight: 1440,
	}

	if err := uc.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	exists, _ := uc.Exists()
	if !exists {
		t.Error("expected true after save")
	}

	got, err := uc.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.CameraResWidth != 2560 {
		t.Errorf("CameraResWidth: got %d, want 2560", got.CameraResWidth)
	}
}

func TestVRChatConfigUseCase_Delete(t *testing.T) {
	repo := &fakeConfigRepo{cfg: &vrchatconfig.VRChatConfig{}}
	uc := NewVRChatConfigUseCase(repo)

	if err := uc.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	exists, _ := uc.Exists()
	if exists {
		t.Error("expected false after delete")
	}
}

func TestVRChatConfigUseCase_Get_NotExist(t *testing.T) {
	repo := &fakeConfigRepo{}
	uc := NewVRChatConfigUseCase(repo)

	_, err := uc.Get()
	if err == nil {
		t.Error("expected error when config does not exist")
	}
}
