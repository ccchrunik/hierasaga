package service

import (
	"atm/ds"
	"fmt"
	"sync/atomic"
)

type RoundGateway struct {
	sys   *System
	queue *ds.MutexQueue
	txNum uint64
}

func NewRoundGateway(sys *System) *RoundGateway {
	return &RoundGateway{
		sys:   sys,
		queue: ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
	}
}

func (rgtw *RoundGateway) Name() string {
	return ServiceGateway
}

func (rgtw *RoundGateway) Send(req Request, round int) {
	rgtw.queue.Push(ds.NewItem(round, req))
}

// gateway is always available
func (rgtw *RoundGateway) Receive() {
	eq := rgtw.sys.EventQueue
	for !rgtw.queue.IsEmpty() {
		item := rgtw.queue.Pop().(*ds.Item)
		req := item.Value().(Request)
		e := rgtw.initEvent(req)
		if e.TxID == "" {
			e.TxID = fmt.Sprintf("tx-%d", atomic.AddUint64(&rgtw.txNum, 1))
		}
		eq.Send(e)
	}
}

func (rgtw *RoundGateway) initEvent(req Request) Event {
	e := NewEvent()
	e.TxID = req.TxID
	e.From = ServiceGateway
	e.To = ServiceTxManager
	e.PushCallStack(req.Service, req.Endpoint, 0)
	e.CurrentRetryTime = 0
	e.RemainingRetryTime = DefaultRetryTime
	e.Round = rgtw.sys.Round() + 1
	e.Phase = PhaseBegin
	return e
}
