package ssm

import (
	"errors"
	"fmt"
)

// State represents a single state of the system.
type State struct {
	// Name is the name of the state, give it something meaningful
	Name string
}

// Trigger is a trigger that can move from one state to another.
type Trigger struct {
	// Key is a unique key for the trigger
	Key string
}

// String returns a useful debug string for the trigger.
func (t Trigger) String() string {
	return fmt.Sprintf(t.Key)
}

// StateMachine is a simple state machine that allows a caller to specify states, triggers
// and edges between states via triggers.  See the examples and test files for more examples.
type StateMachine struct {
	current       *StateConfig
	stateToConfig map[State]*StateConfig
}

// NewStateMachine returns an initialized StateMachine instance.
func NewStateMachine(initial State) *StateMachine {
	sm := &StateMachine{
		stateToConfig: make(map[State]*StateConfig),
	}
	cfg := sm.registerStateConfig(initial)
	sm.current = cfg
	return sm
}

// Configure is used to configure a state, such as setting OnEnter/OnExit
// handlers, defining valid triggers etc.
func (sm *StateMachine) Configure(s State) *StateConfig {
	return sm.registerStateConfig(s)
}

// State returns the current state of the state machine.
func (sm *StateMachine) State() State {
	return sm.current.state
}

// Fire fires the specified trigger. If the trigger is not valid for the current
// state an error is returned.
func (sm *StateMachine) Fire(triggerKey string, ctx interface{}) error {
	if !sm.CanFire(triggerKey) {
		return errors.New("unsupported trigger")
	}

	edge := sm.current.permitted[triggerKey]

	// If the state we are transitioning to is not a substate of the current
	// state then fire all of the exit handlers up the chain
	targetParent := sm.stateToConfig[edge.state].parent
	if targetParent == nil || (targetParent.state != sm.current.state) {
		current := sm.current
		for current != nil {
			if current.onExit != nil {
				current.onExit()
			}
			current = current.parent
		}
	}

	sm.current = sm.stateToConfig[edge.state]

	enterFrom, ok := sm.current.onEnterFrom[edge.trigger]
	if ok {
		enterFrom(ctx)
	}

	if sm.current.onEnter != nil {
		sm.current.onEnter()
	}
	return nil
}

// CanFire returns true if the specified trigger is valid for the State Machines
// current state.
func (sm *StateMachine) CanFire(triggerKey string) bool {
	next, ok := sm.current.permitted[triggerKey]
	if !ok {
		return false
	}

	// If the transtion is guarded, loop to see if we can find a predicate that
	// returns true, indicating this is a valid transition at this time
	if len(next.preds) > 0 {
		found := false
		for _, pred := range next.preds {
			found = pred()
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// IsInState returns true if the current state matches the specified state. It will
// also return true if the current state is a substate of the specified state.
func (sm *StateMachine) IsInState(s State) bool {
	current := sm.current
	for current != nil {
		if current.state == s {
			return true
		}
		current = current.parent
	}
	return false
}

// registerStateConfig registers the state with a blank configuration.
func (sm *StateMachine) registerStateConfig(s State) *StateConfig {
	cfg, ok := sm.stateToConfig[s]
	if !ok {
		cfg = NewStateConfig(sm, s)
		sm.stateToConfig[s] = cfg
	}
	return cfg
}
