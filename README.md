# simple-state-machine
golang based state machine library, for creating simple state machines to handle workflows. Based on the https://github.com/dotnet-state-machine/stateless project.

## Documentation
See [godoc](https://godoc.org/github.com/markdaws/simple-state-machine)

## Installation
```bash
go get github.com/markdaws/simple-state-machine
```

## Package
```go
import "github.com/markdaws/simple-state-machine"
```

## Usage
See state_machine_test.go and examples/main.go for examples on how to use this library.

## State Machine creation
To use the state machine, you need to define all the states, and triggers.  you then create a state machine with an initial state, and configure the state to show how the triggers move between the various states.

```go
off := ssm.State{Name: "off"}
on := ssm.State{Name: "on"}
exited := ssm.State{Name: "exited"}

// each trigger needs a unique key, it can be any string you want
space := ssm.Trigger{Key: " "}
quit := ssm.Trigger{Key: "q"}

// create the state machine initially in the off state
onoff := ssm.NewStateMachine(off)

// You have to configure each state and setup your transitions, 
// configure the off state
cfgOff := onoff.Configure(off)

// Permit specifies which triggers can be fired on a state and 
// what state they will transition to. Here we see on the off state 
//we can fire the "space" trigger and it will go the to "on" state
cfgOff.Permit(space, on)
cfgOff.Permit(quit, exited)

// Next we configure the "on" state, using Permit to specify we can 
// fire the "space" trigger in the "on" state and it will go to 
// the "off" state
cfgOn := onoff.Configure(on)
cfgOn.Permit(space, off)

// If the user enters the exited state then we quit the app
cfgExited := onoff.Configure(exited)
cfgExited.OnEnter(func() { panic("user exited the app") })

reader := bufio.NewReader(os.Stdin)
for {
	fmt.Println("current state: ", onoff.State().Name)
	fmt.Print("Enter text (a single space toggles the state, other strings do nothing): ")
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)

  // Fire the trigger, in this case we pass in the user entered string as the trigger key
  // if it matches any triggers we will transition, otherwise we won't
	err := onoff.Fire(text, nil)
	if err != nil {
		fmt.Println(err)
	}
}
```

## Entry/Exit events
You can specify a handler that will fire when a state is entered and exited.  You can also get more fine grained and specify that a handler should only fire when transitioning to a state from a certain trigger.

```go
s1 := ssm.State{Name: "s1"}
s2 := ssm.State{Name: "s2"}
tr1 := ssm.Trigger{Key: "tr1"}
tr2 := ssm.Trigger{Key: "tr2"}

// Create a new state machine, initially in state s1
sm := ssm.NewStateMachine(s1)

// Configure a handler that fires when we exit s1 and on for when we enter s2
cfg := sm.Configure(s1)

// Hanler that fires when we exit s1
cfg.OnExit(func() { fmt.Println("s1 exit") })

// Allow users to use trigger tr1 to exit s1 and transition to s2 and 
// also trigger tr2 to go from s1 to s2
cfg.Permit(tr1, s2)
cfg.Permit(tr2, s2)

cfg = sm.Configure(s2)
cfg.OnEnter(func() { fmt.Println("s2 enter") })
cfg.OnExit(func() { fmt.Println("s2 exit") })
cfg.OnEnterFrom(tr2, func(ctx interface{}) { fmt.Println("s2 enter from tr2") })
```

## Parameterized Triggers
When you fire a trigger, you can pass along some data that will be passed to the OnEnter from handler if you specified one.
```go
s1 := ssm.State{Name: "s1"}
s2 := ssm.State{Name: "s2"}
tr1 := ssm.Trigger{Key: "tr1"}

sm := ssm.NewStateMachine(s1)
cfg := sm.Configure(s1)
cfg.Permit(tr1, s2)

cfg = sm.Configure(s2)
cfg.OnEnterFrom(tr1, func(ctx interface{}) { fmt.Println("s2 onenterfrom got data", ctx.(string)) })

sm.Fire(tr1.Key, "I am some data")
```

## Guarded Triggers
You can specify a trigger from one state to another, that is only valid if certain conditions are met.
```go
s1 := ssm.State{Name:"s1"}
s2 := ssm.State{Name:"s2"}
tr := ssm.Trigger{Key:"tr"}

sm := ssm.NewStateMachine(s1)
cfg := sm.Configure(s1)

// Using PermitIf vs Permit we can define a predicate that will be checked at 
// runtime to see if we can transition.  You can have multiple predicates, the
// state machine iterates from first to last and will stop at the first predicate
// that evaluates to true
canITransition := false
cfg.PermitIt(tr, s2, func(){ return canITransition })
```

## Substates
You can specify that one state is a substate of another. For example if you have a telephone call, you could be in a "Connected" state but also be in the "OnHold" state, when you are on hold in a call you are still connected, so in this scenario OnHold is considered to be a substate of Connected.

When you are in a substate StateMachine.IsInState will return true if you pass a parent state of the current state, so if the app was in the OnHold state, doing sm.IsInState(Connected) would also return true.

The OnEnter/OnExit handlers behave a little bit differently for substate, when you transition from one state to another, normally the OnExit handler will fire on the state you are leaving, however if you are transitioning to a substate then the OnExit handler will not fire until you leave the substate, for example if S2 is a substate of S1, and S3 is not a substate of either, then if we transition in the following sequence:

S1 -> S2 -> S3

assuming we were already in state S1, then we would see the OnEnter/OnExit handlers fire like:
S2Enter, S2Exit, S1Exit, S3Enter

Notice the S1Exit didn't fire when we moved to S2.  Whereas if S1,S2,S3 were not substate of one another in any way then transitioning

S1 -> S2 -> S3

would produce the events as you expect like:
S1Exit, S2Enter, S2Exit, S3Enter


## Version History
### 0.1.0
Initial release
