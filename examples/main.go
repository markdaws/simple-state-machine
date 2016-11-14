package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/markdaws/simple-state-machine"
)

func main() {
	fmt.Println("Examples")
	fmt.Println(" 1. OnOff")
	fmt.Println(" 2. Bug")
	fmt.Println("")
	fmt.Print("Enter a number:")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	switch text {
	case "1":
		OnOffExample()
	case "2":
		BugExample()
	default:
		fmt.Println("invalid option, good bye!")
	}
}

func OnOffExample() {
	off := ssm.State{Name: "off"}
	on := ssm.State{Name: "on"}
	space := ssm.Trigger{Key: " "}

	onoff := ssm.NewStateMachine(off)
	cfg := onoff.Configure(off)
	cfg.Permit(space, on)
	cfg.OnEnter(func() { fmt.Println("entering off ") })
	cfg.OnExit(func() { fmt.Println("exiting off ") })

	cfg = onoff.Configure(on)
	cfg.Permit(space, off)
	cfg.OnEnter(func() { fmt.Println("entering on ") })
	cfg.OnExit(func() { fmt.Println("exiting on ") })

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("current state: ", onoff.State().Name)
		fmt.Print("Enter text (a single space toggles the state, other strings do nothing): ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		err := onoff.Fire(text, nil)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func BugExample() {
	b := NewBug("bad bug")
	b.Assign("frank")
	b.Assign("joe")
	b.Close()
}

var (
	// states
	Open     = ssm.State{Name: "Open"}
	Assigned = ssm.State{Name: "Assigned"}
	Closed   = ssm.State{Name: "Closed"}

	// triggers
	Assign = ssm.Trigger{Key: "assign"}
	Close  = ssm.Trigger{Key: "close"}
)

type Bug struct {
	Title    string
	Assignee string
	sm       *ssm.StateMachine
}

func NewBug(title string) *Bug {
	b := &Bug{Title: title}

	sm := ssm.NewStateMachine(Open)
	cfg := sm.Configure(Open)
	cfg.Permit(Assign, Assigned)

	cfg = sm.Configure(Assigned)
	cfg.SubstateOf(Open)
	cfg.Permit(Close, Closed)
	cfg.Permit(Assign, Assigned)
	cfg.OnEnterFrom(Assign, func(ctx interface{}) {
		b.Assignee = ctx.(string)
		b.SendEmail(fmt.Sprintf("%s assigned to you", b.Title))
	})
	cfg.OnExit(func() { b.Deassigned() })

	cfg = sm.Configure(Closed)
	cfg.OnEnter(func() {
		b.SendEmail(fmt.Sprintf("%s has been closed", b.Title))
	})

	b.sm = sm
	return b
}

func (b *Bug) Assign(assignee string) {
	err := b.sm.Fire(Assign.Key, assignee)
	if err != nil {
		fmt.Println("assign failed", err)
	}
}

func (b *Bug) Deassigned() {
	b.SendEmail(fmt.Sprintf("%s has been unassigned from you", b.Title))
}

func (b *Bug) Close() {
	err := b.sm.Fire(Close.Key, nil)
	if err != nil {
		fmt.Println("close failed", err)
	}
}

func (b *Bug) SendEmail(msg string) {
	fmt.Printf("Sending Email => %s - %s\n", b.Assignee, msg)
}
