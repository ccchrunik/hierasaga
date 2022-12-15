package service

import "atm/ds"

type NotificationService struct {
	sys        *System
	queue      ds.Queue
	dispatcher *EventDispatcher
}

func NewNotificationService(sys *System) *NotificationService {
	dispatcher := NewEventDispatcher(sys.EventQueue, ServiceNotification)

	dispatcher.Focus("notification").
		Add(func(e Event) (Event, error) {
			return e, nil
		})

	return &NotificationService{
		sys:        sys,
		queue:      ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		dispatcher: dispatcher,
	}
}

func (ns *NotificationService) Name() string {
	return ServiceNotification
}

func (ns *NotificationService) Receive() {
	eq := ns.sys.EventQueue
	for {
		e, err := eq.Pull(ServiceNotification)
		if err != nil {
			break
		}
		ns.dispatcher.Dispatch(e)
	}
}
