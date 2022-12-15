package service

import "fmt"

type EventFunc func(e Event) (Event, error)

type EventFuncChain struct {
	chain []EventFunc
}

type EventDispatcher struct {
	registry map[string]*EventFuncChain
	eq       *EventQueue
	srv      string
}

func NewEventFuncChain() *EventFuncChain {
	return &EventFuncChain{
		chain: []EventFunc{},
	}
}

func (ec *EventFuncChain) Add(ef EventFunc) *EventFuncChain {
	ec.chain = append(ec.chain, ef)
	return ec
}

func (ec *EventFuncChain) Len() int {
	return len(ec.chain)
}

func (ec *EventFuncChain) Select(stage int) EventFunc {
	return ec.chain[stage]
}

func NewEventDispatcher(eq *EventQueue, srv string) *EventDispatcher {
	return &EventDispatcher{
		registry: map[string]*EventFuncChain{},
		eq:       eq,
		srv:      srv,
	}
}

func (ed *EventDispatcher) Focus(endpoint string) *EventFuncChain {
	entry, ok := ed.registry[endpoint]
	if !ok {
		entry = NewEventFuncChain()
		ed.registry[endpoint] = entry
	}
	return entry
}

func (ed *EventDispatcher) Enter(endpoint string, stage int, e Event) (Event, error) {
	chain := ed.registry[endpoint]
	if stage < 0 || stage >= chain.Len() {
		return Event{}, ErrWrongStage
	}
	ef := chain.Select(stage)
	return ef(e)
}

func (ed *EventDispatcher) Dispatch(e Event) {
	newEvent, err := ed.Enter(e.Endpoint, e.Stage, e)
	if err != nil {
		fmt.Printf("unknown dispatch: %v\n", e)
		return
	}
	newEvent.Advance()
	newEvent.From = ed.srv
	// call the child endpoint
	if !newEvent.Equal(&e) {
		newEvent.PushCallStack(e.To, e.Endpoint, e.Stage+1)
	} else {
		// advance the current endpoint stage
		newEvent.Stage++
		// the last stage
		if newEvent.Stage == ed.registry[e.Endpoint].Len() {
			newEvent.Return()
		}
	}
	ed.eq.Send(newEvent)
}
