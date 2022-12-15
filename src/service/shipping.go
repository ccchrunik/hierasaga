package service

import (
	"atm/ds"
)

type ShippingService struct {
	sys        *System
	queue      ds.Queue
	dispatcher *EventDispatcher
}

func NewShippingService(sys *System) *ShippingService {
	dispatcher := NewEventDispatcher(sys.EventQueue, ServiceShipping)

	dispatcher.Focus("shipping").
		Add(func(e Event) (Event, error) {
			return e, nil
		})

	return &ShippingService{
		sys:        sys,
		queue:      ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		dispatcher: dispatcher,
	}
}

func (sps *ShippingService) Name() string {
	return ServiceShipping
}

func (sps *ShippingService) Receive() {
	eq := sps.sys.EventQueue
	for {
		e, err := eq.Pull(ServiceShipping)
		if err != nil {
			break
		}
		sps.dispatcher.Dispatch(e)
	}
}
