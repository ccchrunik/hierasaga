package service

import (
	"atm/ds"
	"fmt"
	"sync/atomic"
)

type RoundGateway struct {
	cfg   *SystemConfig
	queue *ds.MutexQueue
	txNum uint64
}

func NewRoundGateway(cfg *SystemConfig) *RoundGateway {
	return &RoundGateway{
		cfg:   cfg,
		queue: ds.NewMutexTimedPriorityQueue(&cfg.round),
	}
}

func (rgtw *RoundGateway) Name() string {
	return ServiceGateway
}

func (rgtw *RoundGateway) Send(msg Message, round int) {
	rgtw.queue.Push(ds.NewItem(round, msg))
}

func (rgtw *RoundGateway) Receive() {
	services := rgtw.cfg.services
	mq := services[ServiceMessageQueue]
	nextRound := rgtw.cfg.round + 1
	for !rgtw.queue.IsEmpty() {
		msg := rgtw.queue.Pop().(Message)
		// uninitialized message
		if msg.TxID == "" {
			rgtw.initMessage(&msg)
		}
		if msg.RemainingRetryTime == 0 {
			LogErrorMessage(&msg, ErrTooManyRetries)
			continue
		}
		LogMessage(&msg)
		mq.Send(msg, nextRound)
	}
}

func (rgtw *RoundGateway) initMessage(msg *Message) {
	msg.TxID = fmt.Sprintf("%d", atomic.AddUint64(&rgtw.txNum, 1))
	msg.CurrentRetryTime = 0
	msg.RemainingRetryTime = DefaultRetryTime
	msg.Service = ServiceGateway
	msg.NextService = ServiceMessageQueue
	msg.Stack = append(msg.Stack, ServiceGateway)
	msg.Phase = PhasePrepare
	msg.StartRound = rgtw.cfg.round
	msg.CurrentRound = msg.StartRound
	msg.MessageType = TypeRequest
}
