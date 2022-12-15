package service

import (
	"atm/ds"
)

type OrderService struct {
	sys        *System
	queue      ds.Queue
	dispatcher *EventDispatcher
}

func NewOrderService(sys *System) *OrderService {
	dispatcher := NewEventDispatcher(sys.EventQueue, ServiceOrder)

	dispatcher.Focus("order").
		Add(func(e Event) (Event, error) {
			e.To = ServiceShipping
			e.Endpoint = "shipping"
			e.Stage = 0
			return e, nil
		}).
		Add(func(e Event) (Event, error) {
			e.To = ServiceCustomer
			e.Endpoint = "customer"
			e.Stage = 0
			return e, nil
		}).
		Add(func(e Event) (Event, error) {
			return e, nil
		})

	return &OrderService{
		sys:        sys,
		queue:      ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		dispatcher: dispatcher,
	}
}

func (ods *OrderService) Name() string {
	return ServiceOrder
}

func (ods *OrderService) Receive() {
	eq := ods.sys.EventQueue
	for {
		e, err := eq.Pull(ServiceOrder)
		if err != nil {
			break
		}
		ods.dispatcher.Dispatch(e)
	}
}
