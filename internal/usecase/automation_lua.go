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
	runCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()
	return r.runUnsafe(runCtx, source, ev)
}

func (r *scriptRunner) runUnsafe(ctx context.Context, source string, ev automation.Event) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("lua script panicked")
		}
	}()

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()
	L.SetContext(ctx)
	openSafeLuaLibs(L)

	subscriptions := make(map[string][]*lua.LFunction)
	tweaker := L.NewTable()
	L.SetFuncs(tweaker, map[string]lua.LGFunction{
		"subscribe": func(L *lua.LState) int {
			name := L.CheckString(1)
			fn := L.CheckFunction(2)
			subscriptions[name] = append(subscriptions[name], fn)
			return 0
		},
	})
	actions := L.NewTable()
	L.SetField(actions, "run", L.NewFunction(func(L *lua.LState) int {
		actionType := L.CheckString(1)
		payload := luaTableToMap(L.CheckTable(2))
		if err := r.runAction(ctx, actionType, payload); err != nil {
			L.RaiseError("%s", err.Error())
		}
		return 0
	}))
	L.SetField(tweaker, "actions", actions)
	L.SetGlobal("tweaker", tweaker)

	if err := L.DoString(source); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("lua execution timeout")
		}
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
		if ctx.Err() != nil {
			return fmt.Errorf("lua execution timeout")
		}
		L.Push(fn)
		L.Push(lua.LString(ev.Type))
		L.Push(payloadTable)
		if err := L.PCall(2, 0, nil); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("lua execution timeout")
			}
			return err
		}
	}
	return nil
}

func openSafeLuaLibs(L *lua.LState) {
	for _, lib := range []struct {
		name string
		fn   lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(lib.fn))
		L.Push(lua.LString(lib.name))
		L.Call(1, 0)
	}
	// Base opens load/dofile/loadfile; strip FS and arbitrary code loaders.
	for _, name := range []string{"dofile", "loadfile", "load", "loadstring"} {
		L.SetGlobal(name, lua.LNil)
	}
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
	case lua.LTTable:
		if t, ok := v.(*lua.LTable); ok {
			return luaTableToMap(t)
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
	case map[string]interface{}:
		return mapToLuaTable(L, x)
	case []interface{}:
		t := L.NewTable()
		for i, elem := range x {
			L.RawSetInt(t, i+1, goToLua(L, elem))
		}
		return t
	default:
		return lua.LString(fmt.Sprint(v))
	}
}
