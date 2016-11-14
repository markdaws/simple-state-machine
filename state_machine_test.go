package ssm_test

import (
	"testing"

	"github.com/markdaws/simple-state-machine"
	"github.com/stretchr/testify/require"
)

func TestSettingInitialState(t *testing.T) {
	s := ssm.State{Name: "foo"}
	sm := ssm.NewStateMachine(s)
	require.Equal(t, s, sm.State())
}

func TestOnEnterOnExit(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	tr := ssm.Trigger{Key: "foo"}

	sm := ssm.NewStateMachine(s1)
	s1EnterCalled := false
	s1ExitCalled := false
	s2EnterCalled := false
	s2ExitCalled := false

	cfg := sm.Configure(s1)
	cfg.OnEnter(func() { s1EnterCalled = true })
	cfg.OnExit(func() { s1ExitCalled = true })
	cfg.Permit(tr, s2)

	cfg = sm.Configure(s2)
	cfg.OnEnter(func() { s2EnterCalled = true })
	cfg.OnExit(func() { s2ExitCalled = true })

	err := sm.Fire(tr.Key, nil)
	require.Nil(t, err)
	require.False(t, s1EnterCalled)
	require.True(t, s1ExitCalled)
	require.True(t, s2EnterCalled)
	require.False(t, s2ExitCalled)
}

func TestMultiplePermits(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	s3 := ssm.State{Name: "s3"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}

	sm := ssm.NewStateMachine(s1)

	cfg := sm.Configure(s1)
	cfg.Permit(tr1, s2)
	cfg.Permit(tr2, s3)

	err := sm.Fire(tr1.Key, nil)
	require.Nil(t, err)
	require.Equal(t, sm.State(), s2)
}

func TestMultiplePermitsPart2(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	s3 := ssm.State{Name: "s3"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}

	sm := ssm.NewStateMachine(s1)

	cfg := sm.Configure(s1)
	cfg.Permit(tr1, s2)
	cfg.Permit(tr2, s3)

	err := sm.Fire(tr2.Key, nil)
	require.Nil(t, err)
	require.Equal(t, sm.State(), s3)
}

func TestInvalidTrigger(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}

	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s1)
	cfg.Permit(tr1, s2)

	err := sm.Fire(tr2.Key, nil)
	require.NotNil(t, err)
	require.Equal(t, sm.State(), s1)
}

func TestGuardedPermits(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	tr1 := ssm.Trigger{Key: "tr1"}

	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s1)

	allow := false
	cfg.PermitIf(tr1, s2, func() bool { return allow })

	canFire := sm.CanFire(tr1.Key)
	err := sm.Fire(tr1.Key, nil)
	require.False(t, canFire)
	require.NotNil(t, err)
	require.Equal(t, sm.State(), s1)

	allow = true
	canFire = sm.CanFire(tr1.Key)
	err = sm.Fire(tr1.Key, nil)
	require.True(t, canFire)
	require.Nil(t, err)
	require.Equal(t, sm.State(), s2)
}

func TestOnEnterFrom(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	s3 := ssm.State{Name: "s3"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}

	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s3)
	onEnter := false
	onEnterFrom := false
	cfg.OnEnterFrom(tr1, func(ctx interface{}) { onEnterFrom = true })
	cfg.OnEnter(func() { onEnter = true })

	cfg = sm.Configure(s1)
	cfg.Permit(tr1, s3)
	err := sm.Fire(tr1.Key, nil)

	require.Nil(t, err)
	require.True(t, onEnter)
	require.True(t, onEnterFrom)

	sm = ssm.NewStateMachine(s2)
	cfg = sm.Configure(s3)
	onEnter = false
	onEnterFrom = false
	cfg.OnEnterFrom(tr1, func(ctx interface{}) { onEnterFrom = true })
	cfg.OnEnter(func() { onEnter = true })

	cfg = sm.Configure(s2)
	cfg.Permit(tr2, s3)
	err = sm.Fire(tr2.Key, nil)

	require.Nil(t, err)
	require.True(t, onEnter)
	require.False(t, onEnterFrom)
}

func TestFireWithContext(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	tr := ssm.Trigger{Key: "tr"}

	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s1)
	cfg.Permit(tr, s2)

	var onEnterFrom interface{}
	cfg = sm.Configure(s2)
	cfg.OnEnterFrom(tr, func(ctx interface{}) { onEnterFrom = ctx })

	ctx := 12345
	err := sm.Fire(tr.Key, ctx)

	require.Nil(t, err)
	require.Equal(t, ctx, onEnterFrom)
}

func TestReentry(t *testing.T) {
	s1 := ssm.State{Name: "s1"}

	sm := ssm.NewStateMachine(s1)
	tr := ssm.Trigger{Key: "tr"}
	cfg := sm.Configure(s1)

	// A trigger that goes to the same state that we are currently in. In this
	// case the enter/exit handlers should still be fired
	var methods []string
	cfg.Permit(tr, s1)
	cfg.OnExit(func() { methods = append(methods, "exit") })
	cfg.OnEnter(func() { methods = append(methods, "enter") })

	err := sm.Fire(tr.Key, nil)
	require.Nil(t, err)
	require.Equal(t, "exit", methods[0])
	require.Equal(t, "enter", methods[1])
}

func TestIsInState(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	s3 := ssm.State{Name: "s3"}
	s4 := ssm.State{Name: "s4"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}
	tr3 := ssm.Trigger{Key: "tr3"}

	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s1)
	cfg.Permit(tr1, s2)

	cfg = sm.Configure(s2)
	cfg.SubstateOf(s1)
	cfg.Permit(tr2, s3)

	cfg = sm.Configure(s3)
	cfg.SubstateOf(s2)
	cfg.Permit(tr3, s4)

	require.True(t, sm.IsInState(s1))
	err := sm.Fire(tr1.Key, nil)
	require.Nil(t, err)
	require.Equal(t, s2, sm.State())
	require.True(t, sm.IsInState(s2))

	// s2 is a substate of s1, so should still be considered in s1
	require.True(t, sm.IsInState(s1))

	err = sm.Fire(tr2.Key, nil)
	require.Nil(t, err)
	require.Equal(t, s3, sm.State())
	require.True(t, sm.IsInState(s1))
	require.True(t, sm.IsInState(s2))
	require.True(t, sm.IsInState(s3))
	require.False(t, sm.IsInState(s4))

	err = sm.Fire(tr3.Key, nil)
	require.Nil(t, err)
	require.Equal(t, s4, sm.State())
	require.False(t, sm.IsInState(s1))
	require.False(t, sm.IsInState(s2))
	require.False(t, sm.IsInState(s3))
	require.True(t, sm.IsInState(s4))
}

func TestOnEnterOnExitSubstates(t *testing.T) {
	// When entering a substate, the parent state should not fire
	// an OnExit handler since the substate is still considered to
	// be in the parents state
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	s3 := ssm.State{Name: "s3"}
	s4 := ssm.State{Name: "s4"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}
	tr3 := ssm.Trigger{Key: "tr3"}

	var calls []string
	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s1)
	cfg.Permit(tr1, s2)
	cfg.OnEnter(func() { calls = append(calls, "s1enter") })
	cfg.OnExit(func() { calls = append(calls, "s1exit") })

	cfg = sm.Configure(s2)
	cfg.Permit(tr2, s3)
	cfg.OnEnter(func() { calls = append(calls, "s2enter") })
	cfg.OnExit(func() { calls = append(calls, "s2exit") })

	cfg = sm.Configure(s3)
	cfg.SubstateOf(s2)
	cfg.Permit(tr3, s4)
	cfg.OnEnter(func() { calls = append(calls, "s3enter") })
	cfg.OnExit(func() { calls = append(calls, "s3exit") })

	cfg = sm.Configure(s4)
	cfg.OnEnter(func() { calls = append(calls, "s4enter") })
	cfg.OnExit(func() { calls = append(calls, "s4exit") })

	err := sm.Fire(tr1.Key, nil)
	require.Nil(t, err)

	err = sm.Fire(tr2.Key, nil)
	require.Nil(t, err)

	err = sm.Fire(tr3.Key, nil)
	require.Nil(t, err)

	require.Equal(t, []string{"s1exit", "s2enter", "s3enter", "s3exit", "s2exit", "s4enter"}, calls)
}

func TestCanFire(t *testing.T) {
	s1 := ssm.State{Name: "s1"}
	s2 := ssm.State{Name: "s2"}
	s3 := ssm.State{Name: "s3"}
	tr1 := ssm.Trigger{Key: "tr1"}
	tr2 := ssm.Trigger{Key: "tr2"}
	tr3 := ssm.Trigger{Key: "tr3"}

	sm := ssm.NewStateMachine(s1)
	cfg := sm.Configure(s1)
	cfg.Permit(tr1, s2)

	canTransition := false
	cfg.PermitIf(tr3, s3, func() bool { return canTransition })

	require.True(t, sm.CanFire(tr1.Key))
	require.False(t, sm.CanFire(tr2.Key))
	require.False(t, sm.CanFire(tr3.Key))

	canTransition = true
	require.True(t, sm.CanFire(tr3.Key))
}
