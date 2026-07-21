package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/infrastructure/powerplan"

	lua "github.com/yuin/gopher-lua"
)

// PowerPlanService sets the active OS power plan.
type PowerPlanService interface {
	ListDetected(ctx context.Context) ([]powerplan.Plan, error)
	SetActive(ctx context.Context, guid string) error
	ResolvePreset(ctx context.Context, preset string) (string, error)
}

type realPowerPlanService struct{}

func (realPowerPlanService) ListDetected(ctx context.Context) ([]powerplan.Plan, error) {
	return powerplan.ListDetected(ctx)
}

func (realPowerPlanService) SetActive(ctx context.Context, guid string) error {
	return powerplan.SetActive(ctx, guid)
}

func (realPowerPlanService) ResolvePreset(ctx context.Context, preset string) (string, error) {
	return powerplan.ResolvePreset(ctx, preset)
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

	done := make(chan error, 1)
	go func() {
		L := lua.NewState(lua.Options{
			SkipOpenLibs:        true,
			CallStackSize:       64,
			RegistrySize:        256,
			MinimizeStackMemory: true,
		})
		defer L.Close()
		L.SetContext(runCtx)
		done <- r.runUnsafe(runCtx, L, source, ev)
	}()

	select {
	case err := <-done:
		if runCtx.Err() != nil {
			return fmt.Errorf("lua execution timeout")
		}
		return err
	case <-runCtx.Done():
		// SetContext aborts bytecode loops; wait for the VM owner goroutine to Close.
		<-done
		return fmt.Errorf("lua execution timeout")
	}
}

func (r *scriptRunner) runUnsafe(ctx context.Context, L *lua.LState, source string, ev automation.Event) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("automation: lua panic: %v", rec)
			err = fmt.Errorf("lua script panicked")
		}
	}()

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
		if ctx.Err() != nil {
			L.RaiseError("lua execution timeout")
			return 0
		}
		actionType := L.CheckString(1)
		payload, err := luaTableToMap(L.CheckTable(2))
		if err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		// Blocking OS calls observe ctx via CommandContext; wait so we do not leak goroutines.
		done := make(chan error, 1)
		go func() {
			done <- r.runAction(ctx, actionType, payload)
		}()
		select {
		case <-ctx.Done():
			<-done
			L.RaiseError("lua execution timeout")
			return 0
		case err := <-done:
			if err != nil {
				L.RaiseError("%s", err.Error())
			}
			return 0
		}
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
	// Strip FS loaders, module loaders, and stdout helpers from base.
	for _, name := range []string{
		"dofile", "loadfile", "load", "loadstring",
		"collectgarbage", "coroutine", "debug",
		"require", "module", "package", "print",
	} {
		L.SetGlobal(name, lua.LNil)
	}
	// ponytail: gopher-lua SetMx calls os.Exit on OOM — bound string.rep instead.
	if strLib, ok := L.GetGlobal(lua.StringLibName).(*lua.LTable); ok {
		L.SetField(strLib, "rep", L.NewFunction(safeStringRep))
	}
}

func safeStringRep(L *lua.LState) int {
	s := L.CheckString(1)
	n := L.CheckInt(2)
	if n < 0 {
		L.ArgError(2, "negative repeat count")
		return 0
	}
	if n > 0 && len(s) > 0 {
		maxN := automation.MaxLuaStringRepBytes / len(s)
		if n > maxN {
			L.RaiseError("string.rep: result exceeds %d bytes", automation.MaxLuaStringRepBytes)
			return 0
		}
	}
	L.Push(lua.LString(strings.Repeat(s, n)))
	return 1
}

func luaTableToMap(t *lua.LTable) (map[string]interface{}, error) {
	return luaTableToMapVisited(t, map[*lua.LTable]struct{}{})
}

func luaTableToMapVisited(t *lua.LTable, seen map[*lua.LTable]struct{}) (map[string]interface{}, error) {
	if t == nil {
		return nil, nil
	}
	if _, ok := seen[t]; ok {
		return nil, fmt.Errorf("cyclic lua table")
	}
	seen[t] = struct{}{}
	defer delete(seen, t)

	out := make(map[string]interface{})
	var walkErr error
	t.ForEach(func(k, v lua.LValue) {
		if walkErr != nil {
			return
		}
		val, err := luaValueToGo(v, seen)
		if err != nil {
			walkErr = err
			return
		}
		switch key := k.(type) {
		case lua.LString:
			out[string(key)] = val
		case lua.LNumber:
			out[fmt.Sprintf("%v", float64(key))] = val
		default:
			walkErr = fmt.Errorf("unsupported lua table key type %s", k.Type())
		}
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

func luaValueToGo(v lua.LValue, seen map[*lua.LTable]struct{}) (interface{}, error) {
	switch v.Type() {
	case lua.LTNil:
		return nil, nil
	case lua.LTString:
		if s, ok := v.(lua.LString); ok {
			return string(s), nil
		}
	case lua.LTBool:
		if b, ok := v.(lua.LBool); ok {
			return bool(b), nil
		}
	case lua.LTNumber:
		if n, ok := v.(lua.LNumber); ok {
			return float64(n), nil
		}
	case lua.LTTable:
		if t, ok := v.(*lua.LTable); ok {
			return luaTableToMapVisited(t, seen)
		}
	default:
		return v.String(), nil
	}
	return v.String(), nil
}

func mapToLuaTable(L *lua.LState, m map[string]interface{}) *lua.LTable {
	return mapToLuaTableDepth(L, m, 0)
}

const maxGoToLuaDepth = 32

func mapToLuaTableDepth(L *lua.LState, m map[string]interface{}, depth int) *lua.LTable {
	t := L.NewTable()
	if m == nil || depth > maxGoToLuaDepth {
		return t
	}
	for k, v := range m {
		L.SetField(t, k, goToLuaDepth(L, v, depth+1))
	}
	return t
}

func goToLuaDepth(L *lua.LState, v interface{}, depth int) lua.LValue {
	if depth > maxGoToLuaDepth {
		return lua.LNil
	}
	switch x := v.(type) {
	case nil:
		return lua.LNil
	case string:
		return lua.LString(x)
	case bool:
		return lua.LBool(x)
	case float64:
		return lua.LNumber(x)
	case float32:
		return lua.LNumber(x)
	case int:
		return lua.LNumber(x)
	case int64:
		return lua.LNumber(x)
	case map[string]interface{}:
		return mapToLuaTableDepth(L, x, depth)
	case []interface{}:
		t := L.NewTable()
		for i, elem := range x {
			L.RawSetInt(t, i+1, goToLuaDepth(L, elem, depth+1))
		}
		return t
	case []string:
		t := L.NewTable()
		for i, elem := range x {
			L.RawSetInt(t, i+1, lua.LString(elem))
		}
		return t
	default:
		return lua.LString(fmt.Sprint(v))
	}
}
