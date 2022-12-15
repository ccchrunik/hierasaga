package service

import (
	"atm/ds"
	"fmt"
	"sync"
)

type TxManager struct {
	sys   *System
	queue ds.Queue
	// it can be marked by the user
	progress map[string]State
	mu       sync.Mutex
}

func NewTxManager(sys *System) *TxManager {
	return &TxManager{
		sys:      sys,
		queue:    ds.NewMutexTimedPriorityQueue(&sys.Cfg.round),
		progress: map[string]State{},
		mu:       sync.Mutex{},
	}
}

func (tm *TxManager) Name() string {
	return ServiceTxManager
}

func (tm *TxManager) Receive() {
	eq := tm.sys.EventQueue
	for {
		e, err := eq.Pull(ServiceTxManager)
		if err != nil {
			break
		}
		state := tm.getState(e.TxID)
		switch e.Phase {
		case PhaseBegin:
			// nothing start, just discard the message
			if state == StateAbort {
				tm.setState(e.TxID, StateComplete)
				continue
			}
			tm.setState(e.TxID, StateInProgress)
			e.Advance()
			e.Return()
			e.Phase = PhaseProcessing
			e.State = StateNone
			e.From = ServiceTxManager
			eq.Send(e)

		case PhaseProcessing:
			if state == StateAbort {
				tm.rollback(e)
				continue
			}
			if e.State == StateCommit {
				tm.setState(e.TxID, StateCommit)
				e.Advance()
				e.Return()
				e.From = ServiceTxManager
				eq.Send(e)
			} else {
				fmt.Printf("unkwown state: %v\n", e)
			}

		case PhaseEnd:
			if state == StateAbort {
				tm.rollback(e)
				continue
			}
			tm.setState(e.TxID, StateComplete)

		default:
			// Rollback Phase
			fmt.Printf("unkwown phase: %v\n", e)
		}
	}
}

// we use concurrent rollback instead of hierarchical rollback to simplify the implementation
func (tm *TxManager) rollback(e Event) {
	eq := tm.sys.EventQueue
	for {
		newEvent, ok := e.Rollback()
		// empty stack
		if !ok {
			break
		}
		newEvent.Advance()
		eq.Send(newEvent)
	}
}

func (tm *TxManager) getState(txid string) State {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	state, ok := tm.progress[txid]
	if !ok {
		tm.progress[txid] = StateNone
		return StateNone
	}
	return state
}

func (tm *TxManager) setState(txid string, state State) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.progress[txid] = state
}
