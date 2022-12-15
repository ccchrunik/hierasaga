package ds

import (
	"errors"
	"sync"
)

type MutexQueue struct {
	queue Queue
	mu    sync.Mutex
}

func NewMutexQueue(queue Queue) *MutexQueue {
	return &MutexQueue{
		queue: queue,
		mu:    sync.Mutex{},
	}
}

func NewMutexArrayQueue() *MutexQueue {
	return NewMutexQueue(NewArrayQueue())
}

func NewMutexPriorityQueue() *MutexQueue {
	return NewMutexQueue(NewPriorityQueue())
}

func NewMutexTimedPriorityQueue(round *int) *MutexQueue {
	return NewMutexQueue(NewTimedPriorityQueue(round))
}

func (mq *MutexQueue) NewQueue() NewQueueFunc {
	return func() Queue {
		return NewMutexPriorityQueue()
	}
}

func (mq *MutexQueue) Len() int {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return mq.queue.Len()
}

func (mq *MutexQueue) IsEmpty() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return mq.queue.IsEmpty()
}

func (mq *MutexQueue) Clear() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.queue = mq.queue.NewQueue()()
}

func (mq *MutexQueue) Push(v interface{}) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.queue.Push(v)
}

func (mq *MutexQueue) Pop() interface{} {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if mq.queue.IsEmpty() {
		return errors.New("empty queue")
	}
	return mq.queue.Pop()
}

func (mq *MutexQueue) MoveTo(mqDest *MutexQueue) *MutexQueue {
	mq.mu.Lock()
	mqDest.mu.Lock()
	defer mq.mu.Unlock()
	defer mqDest.mu.Unlock()
	mqDest.queue = mq.queue
	mq.queue = mq.queue.NewQueue()()
	return mqDest
}
