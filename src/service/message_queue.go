package service

import (
	"atm/ds"
	"errors"
	"sync"
)

type MessageQueue struct {
	cfg   *SystemConfig
	queue ds.Queue
	mu    sync.Mutex
}

func NewMessageQueue(cfg *SystemConfig) *MessageQueue {
	return &MessageQueue{
		cfg:   cfg,
		queue: ds.NewMutexTimedPriorityQueue(&cfg.round),
		mu:    sync.Mutex{},
	}
}

func (mq *MessageQueue) Name() string {
	return ServiceMessageQueue
}

func (mq *MessageQueue) Send(msg Message, round int) {
	mq.queue.Push(ds.NewItem(round, msg))
}

func (mq *MessageQueue) Receive() {
}

func (mq *MessageQueue) Pull() (Message, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if mq.queue.IsEmpty() {
		return Message{}, errors.New("no message")
	}
	return mq.queue.Pop().(Message), nil
}
