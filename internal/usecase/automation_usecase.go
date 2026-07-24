package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/automation"
)

// StatusSetter sets the user's VRChat status (for change_status action).
type StatusSetter interface {
	SetStatus(ctx context.Context, status string) error
}

// UserDisplayNamer resolves a display name for run log context (no user ID in UI).
type UserDisplayNamer interface {
	DisplayNameFor(ctx context.Context, vrcUserID string) string
}

type automationItemRepo interface {
	List(ctx context.Context) ([]*automation.AutomationItem, error)
	ListEnabled(ctx context.Context) ([]*automation.AutomationItem, error)
	GetByID(ctx context.Context, id string) (*automation.AutomationItem, error)
	Save(ctx context.Context, item *automation.AutomationItem) error
	Delete(ctx context.Context, id string) error
}

var (
	ErrAutomationItemNotFound = automation.ErrItemNotFound
	ErrAutomationInvalidItem  = errors.New("automation item invalid")
)

// AutomationUseCase handles automation items, events, and actions.
type AutomationUseCase struct {
	itemRepo         automationItemRepo
	statusSetter     StatusSetter
	procChecker      automation.VRChatProcessChecker
	powerPlan        PowerPlanService
	windowResizer    VRChatWindowResizer
	displayNamer     UserDisplayNamer
	eventsMu         sync.RWMutex
	events           chan automation.Event
	runLog           *runLogStore
	failLimiter      *failureLogLimiter
	scripts          *scriptRunner
	onRunLogChanged  func()
	runtimeAvailable bool
	shutdown         atomic.Bool
	runCancel        context.CancelFunc

	workerWG    sync.WaitGroup
	schedulerWG sync.WaitGroup
	processWG   sync.WaitGroup
}

// NewAutomationUseCase creates a new AutomationUseCase.
func NewAutomationUseCase(itemRepo automationItemRepo, statusSetter StatusSetter, procChecker automation.VRChatProcessChecker) *AutomationUseCase {
	uc := &AutomationUseCase{
		itemRepo:      itemRepo,
		statusSetter:  statusSetter,
		procChecker:   procChecker,
		powerPlan:     realPowerPlanService{},
		windowResizer: realVRChatWindowResizer{},
		runLog:        newRunLogStore(),
		failLimiter:   newFailureLogLimiter(automation.FailureLogRateLimit),
	}
	uc.scripts = newScriptRunner(uc.runActionStep)
	return uc
}

// SetRunLogChangedHook registers a callback after run log updates.
func (uc *AutomationUseCase) SetRunLogChangedHook(fn func()) {
	uc.onRunLogChanged = fn
}

// Start begins the automation worker, scheduler, and process monitor.
func (uc *AutomationUseCase) Start(ctx context.Context) {
	uc.startPlatform(ctx)
}

// Stop shuts down automation goroutines.
func (uc *AutomationUseCase) Stop() {
	uc.shutdown.Store(true)
	uc.stopPlatform()
}

// RuntimeStatus returns subsystem availability for the UI.
func (uc *AutomationUseCase) RuntimeStatus() automation.RuntimeStatus {
	if !uc.runtimeAvailable {
		return automation.RuntimeStatus{Available: false, ReasonKey: "subsystemUnavailable"}
	}
	return automation.RuntimeStatus{Available: true}
}

// OnFriendJoined enqueues a friend_joined event (live tail only).
func (uc *AutomationUseCase) OnFriendJoined(ctx context.Context, vrcUserID string) error {
	if vrcUserID == "" {
		return nil
	}
	if uc.shutdown.Load() {
		return nil
	}
	ev := automation.Event{
		Type: automation.EventFriendJoined,
		Payload: map[string]interface{}{
			"vrc_user_id": vrcUserID,
		},
	}
	uc.eventsMu.RLock()
	started := uc.events != nil
	uc.eventsMu.RUnlock()
	if !started {
		// Tests / pre-Start: evaluate synchronously.
		uc.handleEvent(ctx, ev)
		return nil
	}
	uc.PublishEvent(ev)
	return nil
}

// ListItems returns all automation items.
func (uc *AutomationUseCase) ListItems(ctx context.Context) ([]*automation.AutomationItem, error) {
	return uc.itemRepo.List(ctx)
}

// GetItem returns one item.
func (uc *AutomationUseCase) GetItem(ctx context.Context, id string) (*automation.AutomationItem, error) {
	return uc.itemRepo.GetByID(ctx, id)
}

// SaveItem validates and persists an item.
func (uc *AutomationUseCase) SaveItem(ctx context.Context, item *automation.AutomationItem) error {
	if err := sanitizeAndValidateAutomationItem(item); err != nil {
		return err
	}
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	return uc.itemRepo.Save(ctx, item)
}

// DeleteItem removes an item.
func (uc *AutomationUseCase) DeleteItem(ctx context.Context, id string) error {
	if err := uc.itemRepo.Delete(ctx, id); err != nil {
		return err
	}
	if uc.failLimiter != nil {
		uc.failLimiter.remove(id)
	}
	return nil
}

// ToggleItem enables or disables an item.
func (uc *AutomationUseCase) ToggleItem(ctx context.Context, id string, enabled bool) error {
	item, err := uc.itemRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	item.IsEnabled = enabled
	return uc.itemRepo.Save(ctx, item)
}

// GetRunLog returns recent run log entries.
func (uc *AutomationUseCase) GetRunLog() []automation.RunLogEntry {
	if uc.runLog == nil {
		return nil
	}
	return uc.runLog.list()
}

// ListDetectedPowerPlans returns OS power plans (empty off Windows).
func (uc *AutomationUseCase) ListDetectedPowerPlans() ([]automation.DetectedPowerPlan, error) {
	if uc.powerPlan == nil {
		return nil, nil
	}
	plans, err := uc.powerPlan.ListDetected(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]automation.DetectedPowerPlan, len(plans))
	for i, p := range plans {
		out[i] = automation.DetectedPowerPlan{GUID: p.GUID, Name: p.Name}
	}
	return out, nil
}

// --- legacy rule API (compat) ---

func (uc *AutomationUseCase) ListRules(ctx context.Context) ([]*automation.AutomationRule, error) {
	items, err := uc.itemRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	var rules []*automation.AutomationRule
	for _, item := range items {
		if item.Kind == automation.KindRule {
			rules = append(rules, itemToLegacyRule(item))
		}
	}
	return rules, nil
}

func (uc *AutomationUseCase) GetRule(ctx context.Context, id string) (*automation.AutomationRule, error) {
	item, err := uc.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrAutomationItemNotFound
	}
	return itemToLegacyRule(item), nil
}

func (uc *AutomationUseCase) SaveRule(ctx context.Context, r *automation.AutomationRule) error {
	item := automation.RuleToItem(r)
	if err := uc.SaveItem(ctx, item); err != nil {
		return err
	}
	r.ID = item.ID
	return nil
}

func (uc *AutomationUseCase) DeleteRule(ctx context.Context, id string) error {
	return uc.DeleteItem(ctx, id)
}

func (uc *AutomationUseCase) ToggleRule(ctx context.Context, id string, enabled bool) error {
	item, err := uc.itemRepo.GetByID(ctx, id)
	if err != nil {
		// Legacy silent no-op for unknown id.
		if errors.Is(err, ErrAutomationItemNotFound) {
			return nil
		}
		return err
	}
	item.IsEnabled = enabled
	return uc.itemRepo.Save(ctx, item)
}

func itemToLegacyRule(item *automation.AutomationItem) *automation.AutomationRule {
	if item == nil || item.Kind != automation.KindRule {
		return nil
	}
	// Legacy API exposes only the first action step.
	steps, _ := automation.ParseActions(item.ActionsJSON)
	r := &automation.AutomationRule{
		ID:            item.ID,
		Name:          item.Name,
		TriggerType:   item.TriggerType,
		ConditionJSON: item.ConditionsJSON,
		IsEnabled:     item.IsEnabled,
	}
	if len(steps) > 0 {
		r.ActionType = steps[0].Type
		if steps[0].Payload != nil {
			b, err := json.Marshal(steps[0].Payload)
			if err == nil {
				r.ActionPayload = string(b)
			}
		}
	}
	return r
}

// sanitizeAndValidateAutomationItem validates an item and normalizes fields
// that must be consistent on disk (e.g. strips incompatible conditions).
func sanitizeAndValidateAutomationItem(item *automation.AutomationItem) error {
	if item == nil {
		return ErrAutomationInvalidItem
	}
	if item.Name == "" {
		return fmt.Errorf("%w: name required", ErrAutomationInvalidItem)
	}
	switch item.Kind {
	case automation.KindRule:
		if item.TriggerType == "" {
			return fmt.Errorf("%w: trigger required", ErrAutomationInvalidItem)
		}
		if item.TriggerType == automation.EventScheduleTick {
			if item.ScheduleJSON == "" {
				return fmt.Errorf("%w: schedule required", ErrAutomationInvalidItem)
			}
			if _, err := automation.ParseSchedule(item.ScheduleJSON); err != nil {
				return fmt.Errorf("%w: %v", ErrAutomationInvalidItem, err)
			}
		} else if item.ScheduleJSON != "" {
			if _, err := automation.ParseSchedule(item.ScheduleJSON); err != nil {
				return fmt.Errorf("%w: %v", ErrAutomationInvalidItem, err)
			}
		}
		conds, err := automation.ParseConditions(item.ConditionsJSON)
		if err != nil {
			return fmt.Errorf("%w: conditions: %v", ErrAutomationInvalidItem, err)
		}
		conds = automation.CompatibleConditions(item.TriggerType, conds)
		b, err := json.Marshal(conds)
		if err != nil {
			return fmt.Errorf("%w: conditions: %v", ErrAutomationInvalidItem, err)
		}
		item.ConditionsJSON = string(b)
		if _, err := automation.ParseActions(item.ActionsJSON); err != nil {
			return fmt.Errorf("%w: actions: %v", ErrAutomationInvalidItem, err)
		}
	case automation.KindScript:
		if len(item.ScriptSource) > automation.MaxScriptBytes {
			return fmt.Errorf("%w: script too large", ErrAutomationInvalidItem)
		}
	default:
		return fmt.Errorf("%w: invalid kind", ErrAutomationInvalidItem)
	}
	return nil
}

func (uc *AutomationUseCase) runChangeStatus(ctx context.Context, payload map[string]interface{}) error {
	if uc.statusSetter == nil || payload == nil {
		return nil
	}
	s, _ := payload["status"].(string)
	if s == "" {
		return nil
	}
	if !allowedStatuses[s] {
		return nil
	}
	return uc.statusSetter.SetStatus(ctx, s)
}

var allowedStatuses = map[string]bool{
	"busy":    true,
	"ask me":  true,
	"join me": true,
}

// EvalAndRun evaluates enabled rules synchronously (tests).
func (uc *AutomationUseCase) EvalAndRun(ctx context.Context, triggerType string, payload map[string]interface{}) error {
	if _, err := uc.itemRepo.ListEnabled(ctx); err != nil {
		return err
	}
	uc.handleEvent(ctx, automation.Event{Type: triggerType, Payload: payload})
	return nil
}

// EvalRules evaluates enabled rules for the given trigger context (tests).
func (uc *AutomationUseCase) EvalRules(ctx context.Context, triggerType string, payload map[string]interface{}) ([]*automation.EvalResult, error) {
	items, err := uc.itemRepo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	evalCtx := uc.buildEvalContext(ctx, automation.Event{Type: triggerType, Payload: payload})
	var results []*automation.EvalResult
	for _, item := range items {
		if item.Kind != automation.KindRule {
			continue
		}
		ok, err := automation.EvalItem(item, evalCtx)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		steps, err := automation.ParseActions(item.ActionsJSON)
		if err != nil {
			return nil, err
		}
		if len(steps) == 0 {
			continue
		}
		results = append(results, &automation.EvalResult{
			ShouldFire:    true,
			ActionType:    steps[0].Type,
			ActionPayload: steps[0].Payload,
		})
	}
	return results, nil
}

// RunActions executes eval results (tests).
func (uc *AutomationUseCase) RunActions(ctx context.Context, results []*automation.EvalResult) error {
	for _, res := range results {
		if res == nil || !res.ShouldFire {
			continue
		}
		if err := uc.runActionStep(ctx, res.ActionType, res.ActionPayload); err != nil {
			return err
		}
	}
	return nil
}
