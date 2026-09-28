package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"vrchat-tweaker/internal/domain/automation"
)

type mockStatusSetter struct {
	mu     sync.Mutex
	called []string
	err    error
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

type mockVRChatProcessChecker struct {
	running bool
	err     error
}

func (m *mockVRChatProcessChecker) VRChatRunning() (bool, error) {
	return m.running, m.err
}

func newTestAutomationUseCase(repo *mockAutomationItemRepo, setter StatusSetter) *AutomationUseCase {
	return NewAutomationUseCase(repo, setter, &mockVRChatProcessChecker{})
}

type mockAutomationItemRepo struct {
	mu          sync.Mutex
	items       []*automation.AutomationItem
	listErr     error
	listEnErr   error
	getByIDErr  error
	saveErr     error
	deleteErr   error
	lastDeleted []string
}

func (m *mockAutomationItemRepo) seedItems(items []*automation.AutomationItem) {
	m.items = append([]*automation.AutomationItem(nil), items...)
}

func testChangeStatusItem(id, trigger, status string, enabled bool) *automation.AutomationItem {
	return &automation.AutomationItem{
		ID:          id,
		Name:        "n",
		Kind:        automation.KindRule,
		TriggerType: trigger,
		IsEnabled:   enabled,
		ActionsJSON: `[{"type":"change_status","payload":{"status":"` + status + `"}}]`,
	}
}

func (m *mockAutomationItemRepo) List(context.Context) ([]*automation.AutomationItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listErr != nil {
		return nil, m.listErr
	}
	return append([]*automation.AutomationItem(nil), m.items...), nil
}

func (m *mockAutomationItemRepo) ListEnabled(context.Context) ([]*automation.AutomationItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listEnErr != nil {
		return nil, m.listEnErr
	}
	var out []*automation.AutomationItem
	for _, it := range m.items {
		if it != nil && it.IsEnabled {
			out = append(out, it)
		}
	}
	return out, nil
}

func (m *mockAutomationItemRepo) GetByID(_ context.Context, id string) (*automation.AutomationItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, it := range m.items {
		if it != nil && it.ID == id {
			cpy := *it
			return &cpy, nil
		}
	}
	return nil, automation.ErrItemNotFound
}

func (m *mockAutomationItemRepo) Save(_ context.Context, item *automation.AutomationItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	for i, it := range m.items {
		if it != nil && it.ID == item.ID {
			cpy := *item
			m.items[i] = &cpy
			return nil
		}
	}
	cpy := *item
	m.items = append(m.items, &cpy)
	return nil
}

func (m *mockAutomationItemRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	found := false
	var next []*automation.AutomationItem
	for _, it := range m.items {
		if it != nil && it.ID == id {
			found = true
			continue
		}
		next = append(next, it)
	}
	if !found {
		return automation.ErrItemNotFound
	}
	m.lastDeleted = append(m.lastDeleted, id)
	m.items = next
	return nil
}

func TestAutomationUseCase_SaveItem_assignsIDWhenEmpty(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	uc := newTestAutomationUseCase(repo, nil)
	item := &automation.AutomationItem{
		Name:        "n",
		Kind:        automation.KindRule,
		TriggerType: automation.TriggerAFKDetected,
		IsEnabled:   true,
		ActionsJSON: `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}
	if err := uc.SaveItem(ctx, item); err != nil {
		t.Fatal(err)
	}
	if item.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestAutomationUseCase_EvalAndRun_propagatesEvalError(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{listEnErr: errors.New("list")}
	uc := newTestAutomationUseCase(repo, nil)
	if err := uc.EvalAndRun(ctx, automation.TriggerAFKDetected, nil); err == nil {
		t.Fatal("want error")
	}
}

func TestAutomationUseCase_EvalAndRun_runsActions(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.seedItems([]*automation.AutomationItem{
		testChangeStatusItem("1", automation.TriggerAFKDetected, "busy", true),
	})
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	if err := uc.EvalAndRun(ctx, automation.TriggerAFKDetected, nil); err != nil {
		t.Fatal(err)
	}
	called := setter.getCalled()
	if len(called) != 1 || called[0] != "busy" {
		t.Fatalf("want [busy], got %v", called)
	}
}

func TestAutomationUseCase_OnFriendJoined(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.seedItems([]*automation.AutomationItem{
		testChangeStatusItem("1", automation.TriggerFriendJoined, "join me", true),
	})
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	if err := uc.OnFriendJoined(ctx, "usr_friend"); err != nil {
		t.Fatal(err)
	}
	called := setter.getCalled()
	if len(called) != 1 || called[0] != "join me" {
		t.Fatalf("want [join me], got %v", called)
	}
	if err := uc.OnFriendJoined(ctx, ""); err != nil {
		t.Fatal(err)
	}
}

func TestAutomationUseCase_OnFriendLeft(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.seedItems([]*automation.AutomationItem{
		testChangeStatusItem("1", automation.TriggerFriendLeft, "ask me", true),
	})
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	if err := uc.OnFriendLeft(ctx, "usr_friend"); err != nil {
		t.Fatal(err)
	}
	called := setter.getCalled()
	if len(called) != 1 || called[0] != "ask me" {
		t.Fatalf("want [ask me], got %v", called)
	}
	if err := uc.OnFriendLeft(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if len(setter.getCalled()) != 1 {
		t.Fatalf("empty user id should not run again, got %v", setter.getCalled())
	}
}

func TestAutomationUseCase_runChangeStatus_edgeCases(t *testing.T) {
	ctx := context.Background()
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, setter)

	if err := uc.runChangeStatus(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if err := uc.runChangeStatus(ctx, map[string]interface{}{"status": "invalid"}); err != nil {
		t.Fatal(err)
	}
	if len(setter.getCalled()) != 0 {
		t.Fatalf("unexpected SetStatus calls: %v", setter.getCalled())
	}

	wantErr := errors.New("api error")
	setter.err = wantErr
	err := uc.runChangeStatus(ctx, map[string]interface{}{"status": "ask me"})
	if err != wantErr {
		t.Errorf("runChangeStatus err = %v, want %v", err, wantErr)
	}
}

func TestAutomationUseCase_ToggleItem_unknownId(t *testing.T) {
	ctx := context.Background()
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, nil)
	if err := uc.ToggleItem(ctx, "missing", true); err == nil {
		t.Fatal("want error")
	}
}
