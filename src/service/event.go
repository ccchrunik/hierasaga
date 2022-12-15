package service

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Event struct {
	TxID               string
	From               string
	To                 string
	Round              int
	CurrentRetryTime   int
	RemainingRetryTime int
	Endpoint           string
	Stage              int
	Phase              Phase
	State              State
	Action             Action
	Controller         string
	CallStack          []string
	RollbackStack      []string
	Body               map[string]interface{}
}

func NewEvent() Event {
	return Event{
		CallStack:     []string{},
		RollbackStack: []string{},
		Body:          map[string]interface{}{},
	}
}

func (e *Event) IsControlledBy(s string) bool {
	return e.Controller == s
}

func (e *Event) PopCallStack() (string, bool) {
	if len(e.CallStack) == 0 {
		return "", false
	}
	lastIndex := len(e.CallStack) - 1
	srv := e.CallStack[lastIndex]
	e.CallStack = e.CallStack[:lastIndex]
	return srv, true
}

func (e *Event) PushCallStack(srv, endpoint string, stage int) {
	e.CallStack = append(e.CallStack, fmt.Sprintf("%s|%s|%d", srv, endpoint, stage))
}

func (e *Event) ClearCallStack() {
	e.CallStack = []string{}
}

func (e *Event) PopRollbackStack() (string, bool) {
	if len(e.RollbackStack) == 0 {
		return "", false
	}
	lastIndex := len(e.RollbackStack) - 1
	srv := e.RollbackStack[lastIndex]
	e.RollbackStack = e.RollbackStack[:lastIndex]
	return srv, true
}

func (e *Event) PushRollbackStack(srv, endpoint string, stage int) {
	e.RollbackStack = append(e.RollbackStack, fmt.Sprintf("%s|%s|%d", srv, endpoint, stage))
}

func (e *Event) Commit() {
	e.PushCallStack(e.To, e.Endpoint, e.Stage+1)
	e.State = StateCommit
	e.To = ServiceTxManager
	e.Endpoint = ""
	e.Stage = 0
}

func (e *Event) Abort() {
	e.PushCallStack(e.To, e.Endpoint, e.Stage+1)
	e.State = StateAbort
	e.To = ServiceTxManager
	e.Endpoint = ""
	e.Stage = 0
}

func (e *Event) End() {
	e.PushCallStack(e.To, e.Endpoint, e.Stage+1)
	e.Phase = PhaseEnd
	e.State = StateNone
	e.To = ServiceTxManager
	e.Endpoint = ""
	e.Stage = 0
}

func (e *Event) Rollback() (Event, bool) {
	newEvent := NewEvent()
	newEvent.TxID = e.TxID
	newEvent.Round = e.Round
	dest, ok := e.PopRollbackStack()
	// no more stack -> done!
	if !ok {
		newEvent.To = ServiceTxManager
		newEvent.Phase = PhaseEnd
		return Event{}, false
	}
	srv, endpoint, stage := ParseDestination(dest)
	newEvent.Phase = PhaseRollback
	newEvent.To = srv
	newEvent.Endpoint = endpoint
	newEvent.Stage, _ = strconv.Atoi(stage)
	return newEvent, true
}

func (e *Event) Return() {
	dest, ok := e.PopCallStack()
	// no more stack -> tx manager
	if !ok {
		e.To = ServiceTxManager
		e.Endpoint = ""
		e.Stage = 0
		e.Phase = PhaseEnd
		return
	}
	srv, endpoint, stage := ParseDestination(dest)
	e.To = srv
	e.Endpoint = endpoint
	e.Stage, _ = strconv.Atoi(stage)
}

func (e *Event) Advance() {
	e.Round++
}

func (e *Event) Get(s string) (interface{}, bool) {
	v, ok := e.Body[s]
	return v, ok
}

func (e *Event) Set(s string, v interface{}) {
	e.Body[s] = v
}

func (e *Event) Delete(s string) {
	delete(e.Body, s)
}

func (e *Event) Equal(e2 *Event) bool {
	return e.To == e2.To && e.Endpoint == e2.Endpoint && e.Stage == e2.Stage
}

func (e *Event) Print() {
}

// dest = Service|Endpoint|Stage
func ParseDestination(dest string) (string, string, string) {
	l := strings.Split(dest, "|")
	return l[0], l[1], l[2]
}

func NextRetryRound(currentRound, retryTime int) int {
	if retryTime <= 0 {
		return -1
	}
	if retryTime > 5 {
		retryTime = 5
	}
	return currentRound + int(math.Exp2(float64(retryTime)))
}
