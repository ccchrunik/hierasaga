package service

import (
	"atm/ds"
)

type RoundGateway struct {
	cfg   *SystemConfig
	qRecv *ds.MutexQueue
	qProc *ds.MutexQueue
}

func NewRoundGateway(cfg *SystemConfig) *RoundGateway {
	return &RoundGateway{
		cfg:   cfg,
		qRecv: ds.NewMutexTimedPriorityQueue(&cfg.round),
		qProc: ds.NewMutexTimedPriorityQueue(&cfg.round),
	}
}

func (rgtw *RoundGateway) Name() string {
	return ServiceGateway
}

// move requests in the receiving from the previous round to the processing queue and clear the sending queue
func (rgtw *RoundGateway) Setup() error {
	rgtw.qRecv.MoveTo(rgtw.qProc)
	return nil
}

func (rgtw *RoundGateway) Send(msg Message) error {
	rgtw.qRecv.Push(ds.NewItem(rgtw.cfg.round, msg))
	return nil
}

func (rgtw *RoundGateway) Execute() error {
	services := rgtw.cfg.services
	for i := 0; i < rgtw.qProc.Len(); i++ {
		msg := rgtw.qProc.Pop().(Message)
		mq := services[ServiceMessageQueue]
		if err := mq.Send(msg); err != nil {
			rgtw.cfg.Log(msg.ServiceFrom, err)
		}
	}
	return nil
}
