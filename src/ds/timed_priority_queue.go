package ds

type TimedQueue struct {
	queue Queue
	round *int
}

func NewTimedPriorityQueue(round *int) *TimedQueue {
	return &TimedQueue{
		queue: NewPriorityQueue(),
		round: round,
	}
}

func (tq *TimedQueue) NewQueue() NewQueueFunc {
	return func() Queue {
		return NewTimedPriorityQueue(tq.round)
	}
}

func (tq *TimedQueue) IsEmpty() bool {
	if tq.queue.IsEmpty() {
		return true
	}
	v, ok := tq.queue.Pop().(*Item)
	if !ok {
		return true
	}
	tq.queue.Push(v)
	return v.timestamp > *tq.round
}

func (tq *TimedQueue) Len() int {
	items := []*Item{}
	for !tq.IsEmpty() {
		items = append(items, tq.queue.Pop().(*Item))
	}
	for _, item := range items {
		tq.Push(item)
	}
	return len(items)
}

func (tq *TimedQueue) Push(v interface{}) {
	tq.queue.Push(v)
}

func (tq *TimedQueue) Pop() interface{} {
	return tq.queue.Pop()
}
