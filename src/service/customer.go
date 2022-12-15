package service

import "atm/ds"

type CustomerService struct {
	sys        *System
	queue      ds.Queue
	dispatcher *EventDispatcher
}

func NewCustomerService(sys *System) *CustomerService {
	dispatcher := NewEventDispatcher(sys.EventQueue, ServiceCustomer)

	dispatcher.Focus("customer").
		Add(func(e Event) (Event, error) {
			return e, nil
		})

	return &CustomerService{
		sys:        sys,
		queue:      ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		dispatcher: dispatcher,
	}
}

func (cts *CustomerService) Name() string {
	return ServiceCustomer
}

func (cts *CustomerService) Receive() {
	eq := cts.sys.EventQueue
	for {
		e, err := eq.Pull(ServiceCustomer)
		if err != nil {
			break
		}
		cts.dispatcher.Dispatch(e)
	}
}
