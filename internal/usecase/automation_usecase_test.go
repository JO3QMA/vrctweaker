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

func (m *mockAutomationItemRepo) seedRules(rules []*automation.AutomationRule) {
	m.items = nil
	for _, r := range rules {
		m.items = append(m.items, automation.RuleToItem(r))
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

func TestAutomationUseCase_SaveRule_assignsIDWhenEmpty(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	uc := newTestAutomationUseCase(repo, nil)
	rule := &automation.AutomationRule{Name: "n", TriggerType: automation.TriggerAFKDetected, IsEnabled: true, ActionType: automation.ActionChangeStatus, ActionPayload: `{"status":"busy"}`}
	if err := uc.SaveRule(ctx, rule); err != nil {
		t.Fatal(err)
	}
	if rule.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestAutomationUseCase_ToggleRule_getFails(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{getByIDErr: errors.New("no")}
	uc := newTestAutomationUseCase(repo, nil)
	if err := uc.ToggleRule(ctx, "x", true); err == nil {
		t.Fatal("want error")
	}
}

func TestAutomationUseCase_ToggleRule_ruleNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	uc := newTestAutomationUseCase(repo, nil)
	if err := uc.ToggleRule(ctx, "missing", true); err != nil {
		t.Fatalf("ToggleRule: got %v, want nil when rule is not found", err)
	}
}

func TestAutomationUseCase_ToggleRule_updates(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.seedRules([]*automation.AutomationRule{{ID: "a", IsEnabled: false, TriggerType: automation.TriggerAFKDetected, ActionType: automation.ActionChangeStatus, ActionPayload: `{"status":"busy"}`}})
	uc := newTestAutomationUseCase(repo, nil)
	if err := uc.ToggleRule(ctx, "a", true); err != nil {
		t.Fatal(err)
	}
	got, err := uc.GetRule(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || !got.IsEnabled {
		t.Fatalf("got %#v", got)
	}
}

func TestAutomationUseCase_EvalRules_filtersByShouldFire(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.seedRules([]*automation.AutomationRule{
		{
			ID:            "1",
			TriggerType:   automation.TriggerAFKDetected,
			ConditionJSON: "",
			ActionType:    automation.ActionChangeStatus,
			ActionPayload: `{"status":"busy"}`,
			IsEnabled:     true,
		},
		{
			ID:            "2",
			TriggerType:   automation.TriggerFriendJoined,
			IsEnabled:     true,
			ActionPayload: `{}`,
		},
	})
	uc := newTestAutomationUseCase(repo, nil)
	results, err := uc.EvalRules(ctx, automation.TriggerAFKDetected, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
}

func TestAutomationUseCase_RunActions_nilStatusSetter(t *testing.T) {
	ctx := context.Background()
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, nil)
	if err := uc.RunActions(ctx, []*automation.EvalResult{{ShouldFire: true, ActionType: automation.ActionChangeStatus, ActionPayload: map[string]interface{}{"status": "busy"}}}); err != nil {
		t.Fatal(err)
	}
}

func TestAutomationUseCase_RunActions_changeStatus(t *testing.T) {
	ctx := context.Background()
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, setter)
	res := &automation.EvalResult{
		ShouldFire:    true,
		ActionType:    automation.ActionChangeStatus,
		ActionPayload: map[string]interface{}{"status": "busy"},
	}
	if err := uc.RunActions(ctx, []*automation.EvalResult{res}); err != nil {
		t.Fatal(err)
	}
	called := setter.getCalled()
	if len(called) != 1 || called[0] != "busy" {
		t.Errorf("SetStatus called %v, want [busy]", called)
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
	repo.seedRules([]*automation.AutomationRule{
		{
			ID:            "1",
			TriggerType:   automation.TriggerAFKDetected,
			ActionType:    automation.ActionChangeStatus,
			ActionPayload: `{"status":"busy"}`,
			IsEnabled:     true,
		},
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
	repo.seedRules([]*automation.AutomationRule{
		{
			ID:            "1",
			TriggerType:   automation.TriggerFriendJoined,
			ActionType:    automation.ActionChangeStatus,
			ActionPayload: `{"status":"join me"}`,
			IsEnabled:     true,
		},
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

func TestAutomationUseCase_ListRules_GetRule_DeleteRule(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	uc := newTestAutomationUseCase(repo, nil)
	r := &automation.AutomationRule{ID: "id1", Name: "n", TriggerType: automation.TriggerAFKDetected, ActionType: automation.ActionChangeStatus, ActionPayload: `{"status":"busy"}`}
	_ = uc.SaveRule(ctx, r)

	list, err := uc.ListRules(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListRules: %v %#v", err, list)
	}
	got, err := uc.GetRule(ctx, "id1")
	if err != nil || got == nil || got.Name != "n" {
		t.Fatalf("GetRule: %v %#v", err, got)
	}
	if err := uc.DeleteRule(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
	list, _ = uc.ListRules(ctx)
	if len(list) != 0 {
		t.Fatalf("after delete: %d", len(list))
	}
}

func TestAutomationUseCase_ToggleItem_unknownId(t *testing.T) {
	ctx := context.Background()
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, nil)
	if err := uc.ToggleItem(ctx, "missing", true); err == nil {
		t.Fatal("want error")
	}
}
