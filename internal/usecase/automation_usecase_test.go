package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/event"
)

type noopEventBus struct{}

func (noopEventBus) Publish(context.Context, string, *event.Event) error { return nil }

func (noopEventBus) Subscribe(string, func(context.Context, *event.Event) error) func() {
	return func() {}
}

type mockAutomationRuleRepo struct {
	mu          sync.Mutex
	rules       []*automation.AutomationRule
	listErr     error
	listEnErr   error
	getByIDErr  error
	saveErr     error
	deleteErr   error
	lastSaved   []*automation.AutomationRule
	lastDeleted []string
}

func (m *mockAutomationRuleRepo) List(context.Context) ([]*automation.AutomationRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listErr != nil {
		return nil, m.listErr
	}
	return append([]*automation.AutomationRule(nil), m.rules...), nil
}

func (m *mockAutomationRuleRepo) ListEnabled(context.Context) ([]*automation.AutomationRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listEnErr != nil {
		return nil, m.listEnErr
	}
	var out []*automation.AutomationRule
	for _, r := range m.rules {
		if r != nil && r.IsEnabled {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *mockAutomationRuleRepo) GetByID(_ context.Context, id string) (*automation.AutomationRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, r := range m.rules {
		if r != nil && r.ID == id {
			cpy := *r
			return &cpy, nil
		}
	}
	return nil, nil
}

func (m *mockAutomationRuleRepo) Save(_ context.Context, r *automation.AutomationRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.lastSaved = append(m.lastSaved, r)
	for i, existing := range m.rules {
		if existing != nil && existing.ID == r.ID {
			cpy := *r
			m.rules[i] = &cpy
			return nil
		}
	}
	cpy := *r
	m.rules = append(m.rules, &cpy)
	return nil
}

func (m *mockAutomationRuleRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.lastDeleted = append(m.lastDeleted, id)
	var next []*automation.AutomationRule
	for _, r := range m.rules {
		if r == nil || r.ID != id {
			next = append(next, r)
		}
	}
	m.rules = next
	return nil
}

type recordingActionRunner struct {
	mu     sync.Mutex
	calls  []*automation.EvalResult
	runErr error
}

func (r *recordingActionRunner) Run(_ context.Context, res *automation.EvalResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, res)
	return r.runErr
}

func TestAutomationUseCase_SaveRule_assignsIDWhenEmpty(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationRuleRepo{}
	uc := NewAutomationUseCase(repo, noopEventBus{}, &recordingActionRunner{})
	rule := &automation.AutomationRule{Name: "n", TriggerType: automation.TriggerAFKDetected, IsEnabled: true}
	if err := uc.SaveRule(ctx, rule); err != nil {
		t.Fatal(err)
	}
	if rule.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestAutomationUseCase_ToggleRule_getFails(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationRuleRepo{getByIDErr: errors.New("no")}
	uc := NewAutomationUseCase(repo, noopEventBus{}, &recordingActionRunner{})
	if err := uc.ToggleRule(ctx, "x", true); err == nil {
		t.Fatal("want error")
	}
}

func TestAutomationUseCase_ToggleRule_updates(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationRuleRepo{
		rules: []*automation.AutomationRule{{ID: "a", IsEnabled: false}},
	}
	uc := NewAutomationUseCase(repo, noopEventBus{}, &recordingActionRunner{})
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
	repo := &mockAutomationRuleRepo{
		rules: []*automation.AutomationRule{
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
		},
	}
	uc := NewAutomationUseCase(repo, noopEventBus{}, &recordingActionRunner{})
	results, err := uc.EvalRules(ctx, automation.TriggerAFKDetected, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
}

func TestAutomationUseCase_RunActions_nilRunner(t *testing.T) {
	ctx := context.Background()
	uc := NewAutomationUseCase(&mockAutomationRuleRepo{}, noopEventBus{}, nil)
	if err := uc.RunActions(ctx, []*automation.EvalResult{{ShouldFire: true}}); err != nil {
		t.Fatal(err)
	}
}

func TestAutomationUseCase_RunActions_invokesRunner(t *testing.T) {
	ctx := context.Background()
	runner := &recordingActionRunner{}
	uc := NewAutomationUseCase(&mockAutomationRuleRepo{}, noopEventBus{}, runner)
	res := &automation.EvalResult{ShouldFire: true, ActionType: automation.ActionChangeStatus}
	if err := uc.RunActions(ctx, []*automation.EvalResult{res}); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("calls: %d", len(runner.calls))
	}
}

func TestAutomationUseCase_EvalAndRun_propagatesEvalError(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationRuleRepo{listEnErr: errors.New("list")}
	uc := NewAutomationUseCase(repo, noopEventBus{}, &recordingActionRunner{})
	if err := uc.EvalAndRun(ctx, automation.TriggerAFKDetected, nil); err == nil {
		t.Fatal("want error")
	}
}

func TestAutomationUseCase_EvalAndRun_runsActions(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationRuleRepo{
		rules: []*automation.AutomationRule{
			{
				ID:            "1",
				TriggerType:   automation.TriggerAFKDetected,
				ActionType:    automation.ActionChangeStatus,
				ActionPayload: `{"status":"busy"}`,
				IsEnabled:     true,
			},
		},
	}
	runner := &recordingActionRunner{}
	uc := NewAutomationUseCase(repo, noopEventBus{}, runner)
	if err := uc.EvalAndRun(ctx, automation.TriggerAFKDetected, nil); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("want 1 run, got %d", len(runner.calls))
	}
}

func TestAutomationUseCase_ListRules_GetRule_DeleteRule(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationRuleRepo{}
	uc := NewAutomationUseCase(repo, noopEventBus{}, &recordingActionRunner{})
	r := &automation.AutomationRule{ID: "id1", Name: "n"}
	_ = repo.Save(ctx, r)

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
