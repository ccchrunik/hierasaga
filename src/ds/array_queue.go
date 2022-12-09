package ds

type ArrayQueue []interface{}

func NewArrayQueue() *ArrayQueue {
	return &ArrayQueue{}
}

func (aq *ArrayQueue) NewQueue() NewQueueFunc {
	return func() Queue {
		return NewArrayQueue()
	}
}

func (aq *ArrayQueue) Len() int {
	return len(*aq)
}

func (aq *ArrayQueue) IsEmpty() bool {
	return len(*aq) == 0
}

func (aq *ArrayQueue) Push(v interface{}) {
	*aq = append(*aq, v)
}

func (aq *ArrayQueue) Pop() interface{} {
	v := (*aq)[0]
	*aq = (*aq)[1:]
	return v
}
