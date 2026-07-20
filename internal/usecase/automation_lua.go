package usecase

import (
	"context"
	"fmt"
	"time"

	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/infrastructure/powerplan"

	lua "github.com/yuin/gopher-lua"
)

// PowerPlanService sets the active OS power plan.
type PowerPlanService interface {
	ListDetected() ([]powerplan.Plan, error)
	SetActive(guid string) error
	ResolvePreset(preset string) (string, error)
}

type realPowerPlanService struct{}

func (realPowerPlanService) ListDetected() ([]powerplan.Plan, error) {
	return powerplan.ListDetected()
}

func (realPowerPlanService) SetActive(guid string) error {
	return powerplan.SetActive(guid)
}

func (realPowerPlanService) ResolvePreset(preset string) (string, error) {
	return powerplan.ResolvePreset(preset)
}

type scriptRunner struct {
	execTimeout time.Duration
	runAction   func(ctx context.Context, actionType string, payload map[string]interface{}) error
}

func newScriptRunner(runAction func(ctx context.Context, actionType string, payload map[string]interface{}) error) *scriptRunner {
	return &scriptRunner{
		execTimeout: automation.LuaExecTimeout,
		runAction:   runAction,
	}
}

func (r *scriptRunner) run(ctx context.Context, source string, ev automation.Event) error {
	if len(source) > automation.MaxScriptBytes {
		return fmt.Errorf("script exceeds %d bytes", automation.MaxScriptBytes)
	}
	done := make(chan error, 1)
	go func() {
		done <- r.runUnsafe(source, ev)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	case <-time.After(r.execTimeout):
		return fmt.Errorf("lua execution timeout")
	}
}

func (r *scriptRunner) runUnsafe(source string, ev automation.Event) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("lua panic: %v", rec)
		}
	}()
	L := lua.NewState(lua.Options{SkipOpenLibs: false})
	defer L.Close()

	subscriptions := make(map[string][]*lua.LFunction)

	L.PreloadModule("tweaker", func(L *lua.LState) int {
		t := L.NewTable()
		L.SetFuncs(t, map[string]lua.LGFunction{
			"subscribe": func(L *lua.LState) int {
				name := L.CheckString(1)
				fn := L.CheckFunction(2)
				subscriptions[name] = append(subscriptions[name], fn)
				return 0
			},
			"actions": func(L *lua.LState) int {
				// placeholder table; run set below
				return 0
			},
		})
		actions := L.NewTable()
		L.SetField(actions, "run", L.NewFunction(func(L *lua.LState) int {
			actionType := L.CheckString(1)
			payload := luaTableToMap(L.CheckTable(2))
			err := r.runAction(context.Background(), actionType, payload)
			if err != nil {
				L.RaiseError("%s", err.Error())
			}
			return 0
		}))
		L.SetField(t, "actions", actions)
		L.Push(t)
		return 1
	})

	if err := L.DoString(`tweaker = require("tweaker")`); err != nil {
		return err
	}
	if err := L.DoString(source); err != nil {
		return err
	}

	refs := subscriptions[ev.Type]
	if fn := L.GetGlobal("on_event"); fn != lua.LNil {
		if f, ok := fn.(*lua.LFunction); ok {
			refs = append(refs, f)
		}
	}
	if len(refs) == 0 {
		return nil
	}
	payloadTable := mapToLuaTable(L, ev.Payload)
	for _, fn := range refs {
		L.Push(fn)
		L.Push(lua.LString(ev.Type))
		L.Push(payloadTable)
		if err := L.PCall(2, 0, nil); err != nil {
			return err
		}
	}
	return nil
}

func luaTableToMap(t *lua.LTable) map[string]interface{} {
	out := make(map[string]interface{})
	t.ForEach(func(k, v lua.LValue) {
		key, ok := k.(lua.LString)
		if !ok {
			return
		}
		out[string(key)] = luaValueToGo(v)
	})
	return out
}

func luaValueToGo(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTString:
		if s, ok := v.(lua.LString); ok {
			return string(s)
		}
	case lua.LTBool:
		if b, ok := v.(lua.LBool); ok {
			return bool(b)
		}
	case lua.LTNumber:
		if n, ok := v.(lua.LNumber); ok {
			return float64(n)
		}
	default:
		return v.String()
	}
	return v.String()
}

func mapToLuaTable(L *lua.LState, m map[string]interface{}) *lua.LTable {
	t := L.NewTable()
	if m == nil {
		return t
	}
	for k, v := range m {
		L.SetField(t, k, goToLua(L, v))
	}
	return t
}

func goToLua(L *lua.LState, v interface{}) lua.LValue {
	switch x := v.(type) {
	case string:
		return lua.LString(x)
	case bool:
		return lua.LBool(x)
	case float64:
		return lua.LNumber(x)
	case int:
		return lua.LNumber(x)
	default:
		return lua.LString(fmt.Sprint(v))
	}
}
