package ssm

// edge represents a transition edge from one state to another via a trigger.
type edge struct {
	trigger Trigger
	state   State
	preds   []func() bool
}

// StateConfig stores all of the config information for a state.
type StateConfig struct {
	owner       *StateMachine
	onEnter     func()
	onEnterFrom map[Trigger]func(interface{})
	onExit      func()
	state       State
	parent      *StateConfig
	permitted   map[string]*edge
}

// NewStateConfig returns an initialized StateConfig instance
func NewStateConfig(sm *StateMachine, s State) *StateConfig {
	return &StateConfig{
		owner:       sm,
		state:       s,
		onEnterFrom: make(map[Trigger]func(interface{})),
		permitted:   make(map[string]*edge),
	}
}

// Permit adds an entry indicating the state is permitted to transition to the target
// state via the specified trigger.
func (c *StateConfig) Permit(t Trigger, s State) *StateConfig {
	c.owner.registerStateConfig(s)
	c.permitted[t.Key] = &edge{trigger: t, state: s}
	return c
}

// PermitIf defines a relationship from one state to another via a trigger, which is valid
// when the predicate function evaluates to true.  You can use this to say that we can transition
// from one state to another via a trigger only under certain conditions
func (c *StateConfig) PermitIf(t Trigger, s State, pred func() bool) *StateConfig {
	c.owner.registerStateConfig(s)

	val, ok := c.permitted[t.Key]
	if !ok {
		val = &edge{trigger: t, state: s}
		c.permitted[t.Key] = val
	}

	val.preds = append(val.preds, pred)
	return c
}

// OnEnter registers a handler that will be fired when the state is entered.  This is also fired for
// re-entrant transitions where a state transitions to itself. This handler is called for all triggers
// that enter a state, if you only want to perform an action entering a state for a certain trigger, then
// use OnEnterFrom instead.
func (c *StateConfig) OnEnter(f func()) *StateConfig {
	c.onEnter = f
	return c
}

// OnEnterFrom registers a handler that will fire when entering a state only via the specified trigger.
func (c *StateConfig) OnEnterFrom(t Trigger, f func(interface{})) *StateConfig {
	c.onEnterFrom[t] = f
	return c
}

// OnExit registers a handler that will fire when we exit a state. This will also for for re-entrant transitions
// where we transition from a state to itself.
func (c *StateConfig) OnExit(f func()) *StateConfig {
	c.onExit = f
	return c
}

// SubstateOf specifies that a state is a substate of another.  This means you can specify that state B is a substate
// of state A and if the state machine is currently in state B, asking IsInState(A) will return true ans well as IsInState(B).
// This is also true for any depth of substate relationship.
func (c *StateConfig) SubstateOf(s State) *StateConfig {
	cfg := c.owner.registerStateConfig(s)
	c.parent = cfg
	return c
}
