package service

import "atm/ds"

type PaymentService struct {
	sys        *System
	queue      ds.Queue
	dispatcher *EventDispatcher
}

func NewPaymentService(sys *System) *PaymentService {
	dispatcher := NewEventDispatcher(sys.EventQueue, ServicePayment)

	dispatcher.Focus("payment_control").
		Add(func(e Event) (Event, error) {
			e.To = ServiceOrder
			e.Endpoint = "order"
			e.Stage = 0
			return e, nil
		}).
		Add(func(e Event) (Event, error) {
			e.Commit()
			return e, nil
		}).
		Add(func(e Event) (Event, error) {
			e.To = ServicePayment
			e.Endpoint = "payment_data"
			e.Stage = 0
			return e, nil
		}).
		Add(func(e Event) (Event, error) {
			e.End()
			return e, nil
		})

	dispatcher.Focus("payment_data").
		Add(func(e Event) (Event, error) {
			e.To = ServiceNotification
			e.Endpoint = "notification"
			e.Stage = 0
			return e, nil
		}).
		Add(func(e Event) (Event, error) {
			return e, nil
		})
	return &PaymentService{
		sys:        sys,
		queue:      ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		dispatcher: dispatcher,
	}
}

func (ps *PaymentService) Name() string {
	return ServicePayment
}

func (ps *PaymentService) Receive() {
	eq := ps.sys.EventQueue
	for {
		e, err := eq.Pull(ServicePayment)
		if err != nil {
			break
		}
		ps.dispatcher.Dispatch(e)
	}
}
