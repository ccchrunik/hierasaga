package service

import (
	"atm/ds"
	"strconv"
	"sync"
)

type Executor struct {
	cfg   *SystemConfig
	queue ds.Queue
}

func NewExecutor(cfg *SystemConfig) *Executor {
	return &Executor{
		cfg:   cfg,
		queue: ds.NewMutexTimedPriorityQueue(&cfg.round),
	}
}

func (ec *Executor) Name() string {
	return ServiceExecutor
}

func (ec *Executor) Send(msg Message, round int) {
	ec.queue.Push(ds.NewItem(round, msg))
}

func (ec *Executor) Receive() error {
	services := ec.cfg.services
	nextRound := ec.cfg.round + 1
	mq := services[ServiceMessageQueue].(*MessageQueue)

	var wg sync.WaitGroup
	wg.Add(2)
	// handle return message
	go func() {
		defer wg.Done()
		for !ec.queue.IsEmpty() {
			msg := ec.queue.Pop().(Message)
			msg.CurrentRound++
			LogMessage(&msg)
			switch msg.MessageType {
			case TypeResponse:
				msg.CurrentRetryTime = 0
				msg.RemainingRetryTime = DefaultRetryTime
				switch msg.Phase {
				case PhaseEnd:
					mq.Send(msg, nextRound)

				default:
					dest, ok := msg.PopStack()
					if !ok {
						LogErrorMessage(&msg, ErrNoNmoreService)
					}

					nextSrv, endpoint, stage := ParseDestination(dest)
					msg.NextService = nextSrv
					msg.Endpoint = endpoint
					stageInt, _ := strconv.Atoi(stage)
					msg.Stage = stageInt

					mq.Send(msg, nextRound)
				}

			default:
				LogErrorMessage(&msg, ErrWrongMessageType)
			}
		}
	}()

	// handle transaction messages
	go func() {
		defer wg.Done()
		for {
			msg, err := mq.Pull()
			// no message
			if err != nil {
				break
			}
			msg.CurrentRound++
			LogMessage(&msg)
			switch msg.MessageType {
			case TypeRequest:
				// TODO: now we bypass the check of the transaction manager
				switch msg.Phase {
				case PhaseEnd:
					LogDoneMessage(&msg)
				default:
					dest := services[msg.NextService]
					dest.Send(msg, nextRound)
				}
			default:
				LogErrorMessage(&msg, ErrWrongMessageType)
			}
		}
	}()
	wg.Wait()

	return nil
}
