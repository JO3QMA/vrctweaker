package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/infrastructure/vrchatwindow"
)

var errUnsupportedPowerPlan = fmt.Errorf("set_power_plan: unsupported platform")

func (uc *AutomationUseCase) startPlatform(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	uc.eventsMu.Lock()
	uc.shutdown.Store(false)
	uc.runCancel = cancel
	uc.events = make(chan automation.Event, automation.EventQueueCapacity)
	uc.eventsMu.Unlock()

	if uc.runLog == nil {
		uc.runLog = newRunLogStore()
	}
	if uc.failLimiter == nil {
		uc.failLimiter = newFailureLogLimiter(automation.FailureLogRateLimit)
	}
	if uc.scripts == nil {
		uc.scripts = newScriptRunner(uc.runActionStep)
	}
	uc.runtimeAvailable = true

	uc.workerWG.Add(1)
	go uc.workerLoop(runCtx)

	uc.schedulerWG.Add(1)
	go uc.schedulerLoop(runCtx)

	uc.processWG.Add(1)
	go uc.processMonitorLoop(runCtx)
}

func (uc *AutomationUseCase) stopPlatform() {
	uc.shutdown.Store(true)
	uc.eventsMu.Lock()
	if uc.runCancel != nil {
		uc.runCancel()
		uc.runCancel = nil
	}
	ch := uc.events
	uc.events = nil
	if ch != nil {
		close(ch)
	}
	uc.eventsMu.Unlock()
	uc.workerWG.Wait()
	uc.schedulerWG.Wait()
	uc.processWG.Wait()
}

func (uc *AutomationUseCase) PublishEvent(ev automation.Event) {
	// Hold RLock across send so stopPlatform cannot close mid-send.
	uc.eventsMu.RLock()
	defer uc.eventsMu.RUnlock()
	if uc.shutdown.Load() || uc.events == nil {
		return
	}
	select {
	case uc.events <- ev:
	default:
		log.Printf("automation: event queue full, dropping %s", ev.Type)
	}
}

func (uc *AutomationUseCase) workerLoop(ctx context.Context) {
	defer uc.workerWG.Done()
	uc.eventsMu.RLock()
	ch := uc.events
	uc.eventsMu.RUnlock()
	if ch == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						log.Printf("automation: worker panic: %v", rec)
					}
				}()
				uc.handleEvent(ctx, ev)
			}()
		}
	}
}

func (uc *AutomationUseCase) handleEvent(ctx context.Context, ev automation.Event) {
	items, err := uc.itemRepo.ListEnabled(ctx)
	if err != nil {
		log.Printf("automation: list enabled: %v", err)
		return
	}
	items = automation.SortItemsByID(items)
	evalCtx := uc.buildEvalContext(ctx, ev)
	for _, item := range items {
		uc.handleItem(ctx, item, ev, evalCtx)
	}
}

func (uc *AutomationUseCase) buildEvalContext(ctx context.Context, ev automation.Event) *automation.EvalContext {
	ec := &automation.EvalContext{
		TriggerType: ev.Type,
		Payload:     ev.Payload,
	}
	if uc.procChecker != nil {
		running, err := uc.procChecker.VRChatRunning()
		if err == nil {
			ec.VRChatRunningOK = true
			ec.VRChatRunning = running
		}
	}
	return ec
}

func (uc *AutomationUseCase) handleItem(ctx context.Context, item *automation.AutomationItem, ev automation.Event, evalCtx *automation.EvalContext) {
	now := time.Now()
	ctxLabel := uc.contextLabel(ctx, ev)
	switch item.Kind {
	case automation.KindScript:
		if err := uc.scripts.run(ctx, item.ScriptSource, ev); err != nil {
			uc.recordFailure(item, ev, 0, 0, ctxLabel, err, now)
		} else {
			uc.recordSuccess(item, ev, 0, 0, ctxLabel, now)
		}
	case automation.KindRule:
		ok, err := automation.EvalItem(item, evalCtx)
		if err != nil || !ok {
			return
		}
		completed, total, runErr := uc.runItemActions(ctx, item)
		if runErr != nil {
			uc.recordFailure(item, ev, completed, total, ctxLabel, runErr, now)
		} else {
			uc.recordSuccess(item, ev, completed, total, ctxLabel, now)
		}
	default:
		return
	}
}

func (uc *AutomationUseCase) contextLabel(ctx context.Context, ev automation.Event) string {
	if ev.Type != automation.EventFriendJoined || uc.displayNamer == nil || ev.Payload == nil {
		return ""
	}
	id, _ := ev.Payload["vrc_user_id"].(string)
	if id == "" {
		return ""
	}
	return uc.displayNamer.DisplayNameFor(ctx, id)
}

func (uc *AutomationUseCase) runItemActions(ctx context.Context, item *automation.AutomationItem) (completed, total int, err error) {
	steps, err := automation.ParseActions(item.ActionsJSON)
	if err != nil {
		return 0, 0, err
	}
	total = len(steps)
	for i, step := range steps {
		if err := uc.runActionStep(ctx, step.Type, step.Payload); err != nil {
			if step.ContinueOnError {
				continue
			}
			return i, total, err
		}
		completed = i + 1
	}
	return completed, total, nil
}

func (uc *AutomationUseCase) runActionStep(ctx context.Context, actionType string, payload map[string]interface{}) error {
	switch actionType {
	case automation.ActionChangeStatus:
		return uc.runChangeStatus(ctx, payload)
	case automation.ActionSetPowerPlan:
		return uc.runSetPowerPlan(ctx, payload)
	case automation.ActionSetVRChatWindowSize:
		return uc.runSetVRChatWindowSize(payload)
	default:
		return fmt.Errorf("unknown action %q", actionType)
	}
}

func (uc *AutomationUseCase) runSetPowerPlan(ctx context.Context, payload map[string]interface{}) error {
	if uc.powerPlan == nil {
		return errUnsupportedPowerPlan
	}
	if payload == nil {
		return fmt.Errorf("set_power_plan: empty payload")
	}
	if preset, _ := payload["preset"].(string); preset != "" {
		guid, err := uc.powerPlan.ResolvePreset(ctx, preset)
		if err != nil {
			return err
		}
		return uc.powerPlan.SetActive(ctx, guid)
	}
	if guid, _ := payload["guid"].(string); guid != "" {
		return uc.powerPlan.SetActive(ctx, guid)
	}
	return fmt.Errorf("set_power_plan: preset or guid required")
}

// VRChatWindowResizer resizes the running VRChat client window.
type VRChatWindowResizer interface {
	Resize(width, height int) error
}

type realVRChatWindowResizer struct{}

func (realVRChatWindowResizer) Resize(width, height int) error {
	return vrchatwindow.Resize(width, height)
}

func (uc *AutomationUseCase) runSetVRChatWindowSize(payload map[string]interface{}) error {
	if uc.windowResizer == nil {
		return vrchatwindow.ErrUnsupported
	}
	w, h, err := parseWindowSizePayload(payload)
	if err != nil {
		return err
	}
	return uc.windowResizer.Resize(w, h)
}

func parseWindowSizePayload(payload map[string]interface{}) (width, height int, err error) {
	if payload == nil {
		return 0, 0, fmt.Errorf("set_vrchat_window_size: empty payload")
	}
	w, okW := payloadInt(payload, "width")
	h, okH := payloadInt(payload, "height")
	if !okW || !okH {
		return 0, 0, fmt.Errorf("set_vrchat_window_size: width and height required")
	}
	if w <= 0 || h <= 0 {
		return 0, 0, vrchatwindow.ErrInvalidSize
	}
	return w, h, nil
}

func payloadInt(payload map[string]interface{}, key string) (int, bool) {
	v, ok := payload[key]
	if !ok || v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		if n != float64(int(n)) {
			return 0, false
		}
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	default:
		return 0, false
	}
}

func (uc *AutomationUseCase) recordSuccess(item *automation.AutomationItem, ev automation.Event, completed, total int, ctxLabel string, at time.Time) {
	if total == 0 {
		total = completed
	}
	uc.appendRunLog(automation.RunLogEntry{
		At:               at.UTC().Format(time.RFC3339),
		ItemID:           item.ID,
		ItemName:         item.Name,
		EventType:        ev.Type,
		Success:          true,
		ActionsCompleted: completed,
		ActionsTotal:     total,
		ContextLabel:     ctxLabel,
	})
}

func (uc *AutomationUseCase) recordFailure(item *automation.AutomationItem, ev automation.Event, completed, total int, ctxLabel string, err error, at time.Time) {
	if err == nil {
		err = fmt.Errorf("unknown error")
	}
	if uc.failLimiter != nil && uc.failLimiter.allow(item.ID, at) {
		log.Printf("automation: item %q event %s: %v", item.Name, ev.Type, err)
	}
	if total == 0 {
		total = completed
	}
	uc.appendRunLog(automation.RunLogEntry{
		At:               at.UTC().Format(time.RFC3339),
		ItemID:           item.ID,
		ItemName:         item.Name,
		EventType:        ev.Type,
		Success:          false,
		ActionsCompleted: completed,
		ActionsTotal:     total,
		ContextLabel:     ctxLabel,
		ErrorSummary:     err.Error(),
	})
}

func (uc *AutomationUseCase) appendRunLog(e automation.RunLogEntry) {
	uc.runLog.append(e)
	if uc.onRunLogChanged != nil {
		uc.onRunLogChanged()
	}
}

func (uc *AutomationUseCase) schedulerLoop(ctx context.Context) {
	defer uc.schedulerWG.Done()
	ticker := time.NewTicker(automation.ScheduleTickResolution)
	defer ticker.Stop()
	var lastKey string
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						log.Printf("automation: scheduler panic: %v", rec)
					}
				}()
				key := t.Format("2006-01-02T15:04")
				if key == lastKey {
					return
				}
				lastKey = key
				uc.fireScheduleTick(ctx, t)
			}()
		}
	}
}

func (uc *AutomationUseCase) fireScheduleTick(ctx context.Context, t time.Time) {
	items, err := uc.itemRepo.ListEnabled(ctx)
	if err != nil {
		log.Printf("automation: schedule list: %v", err)
		return
	}
	for _, item := range items {
		if item.Kind != automation.KindRule || item.TriggerType != automation.EventScheduleTick {
			continue
		}
		sched, err := automation.ParseSchedule(item.ScheduleJSON)
		if err != nil {
			continue
		}
		if automation.ScheduleMatches(sched, t) {
			// One schedule.tick per minute; handleEvent evaluates all matching items.
			uc.PublishEvent(automation.Event{Type: automation.EventScheduleTick, Payload: nil})
			return
		}
	}
}

func (uc *AutomationUseCase) processMonitorLoop(ctx context.Context) {
	defer uc.processWG.Done()
	if uc.procChecker == nil {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var (
		hasLast     bool
		lastRunning bool
	)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						log.Printf("automation: process monitor panic: %v", rec)
					}
				}()
				running, err := uc.procChecker.VRChatRunning()
				if err != nil {
					return
				}
				if hasLast && lastRunning == running {
					return
				}
				hasLast = true
				lastRunning = running
				state := "stopped"
				if running {
					state = "running"
				}
				uc.PublishEvent(automation.Event{
					Type: automation.EventVRChatProcess,
					Payload: map[string]interface{}{
						"state": state,
					},
				})
			}()
		}
	}
}
