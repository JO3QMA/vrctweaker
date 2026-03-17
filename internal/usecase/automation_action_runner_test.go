package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"vrchat-tweaker/internal/domain/automation"
)

type mockStatusSetter struct {
	mu      sync.Mutex
	called  []string
	err     error
}

func (m *mockStatusSetter) SetStatus(ctx context.Context, status string) error {
	m.mu.Lock()
	m.called = append(m.called, status)
	err := m.err
	m.mu.Unlock()
	return err
}

func (m *mockStatusSetter) getCalled() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.called...)
}

func TestDefaultActionRunner_Run_ChangeStatus_CallsSetStatus(t *testing.T) {
	m := &mockStatusSetter{}
	r := NewDefaultActionRunner(m)
	ctx := context.Background()

	res := &automation.EvalResult{
		ShouldFire:    true,
		ActionType:    automation.ActionChangeStatus,
		ActionPayload: map[string]interface{}{"status": "busy"},
	}

	err := r.Run(ctx, res)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	called := m.getCalled()
	if len(called) != 1 || called[0] != "busy" {
		t.Errorf("SetStatus called %v, want [busy]", called)
	}
}

func TestDefaultActionRunner_Run_ChangeStatus_IgnoresInvalidStatus(t *testing.T) {
	m := &mockStatusSetter{}
	r := NewDefaultActionRunner(m)
	ctx := context.Background()

	res := &automation.EvalResult{
		ShouldFire:    true,
		ActionType:    automation.ActionChangeStatus,
		ActionPayload: map[string]interface{}{"status": "invalid"},
	}

	_ = r.Run(ctx, res)
	called := m.getCalled()
	if len(called) != 0 {
		t.Errorf("SetStatus should not be called for invalid status, got %v", called)
	}
}

func TestDefaultActionRunner_Run_NilResult_NoOp(t *testing.T) {
	m := &mockStatusSetter{}
	r := NewDefaultActionRunner(m)
	ctx := context.Background()

	_ = r.Run(ctx, nil)
	called := m.getCalled()
	if len(called) != 0 {
		t.Errorf("SetStatus should not be called for nil result, got %v", called)
	}
}

func TestDefaultActionRunner_Run_ShouldFireFalse_NoOp(t *testing.T) {
	m := &mockStatusSetter{}
	r := NewDefaultActionRunner(m)
	ctx := context.Background()

	res := &automation.EvalResult{
		ShouldFire:    false,
		ActionType:    automation.ActionChangeStatus,
		ActionPayload: map[string]interface{}{"status": "busy"},
	}

	_ = r.Run(ctx, res)
	called := m.getCalled()
	if len(called) != 0 {
		t.Errorf("SetStatus should not be called when ShouldFire=false, got %v", called)
	}
}

func TestDefaultActionRunner_Run_ChangeStatus_PropagatesError(t *testing.T) {
	wantErr := errors.New("api error")
	m := &mockStatusSetter{err: wantErr}
	r := NewDefaultActionRunner(m)
	ctx := context.Background()

	res := &automation.EvalResult{
		ShouldFire:    true,
		ActionType:    automation.ActionChangeStatus,
		ActionPayload: map[string]interface{}{"status": "ask me"},
	}

	err := r.Run(ctx, res)
	if err != wantErr {
		t.Errorf("Run err = %v, want %v", err, wantErr)
	}
}
