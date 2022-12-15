package service

import (
	"atm/ds"
)

type EventQueue struct {
	sys    *System
	queues map[string]ds.Queue
}

func NewEventQueue(sys *System) *EventQueue {
	queues := map[string]ds.Queue{
		ServiceTxManager:    ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		ServicePayment:      ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		ServiceOrder:        ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		ServiceShipping:     ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		ServiceCustomer:     ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		ServiceNotification: ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
	}

	return &EventQueue{
		sys:    sys,
		queues: queues,
	}
}

func (eq *EventQueue) Name() string {
	return ServiceEventQueue
}

func (eq *EventQueue) Send(e Event) {
	eq.sys.Log(e.To, e)
	eq.queues[e.To].Push(ds.NewItem(e.Round, e))
}

func (eq *EventQueue) Pull(srv string) (Event, error) {
	// fmt.Printf("srv: %s - clock: %d\n", srv, *eq.clock)
	item, ok := eq.queues[srv].Pop().(*ds.Item)
	if !ok {
		return Event{}, ErrEmptyQueue
	}
	e := item.Value().(Event)
	return e, nil
}
