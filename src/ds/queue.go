package ds

type Queue interface {
	Len() int
	IsEmpty() bool
	Push(interface{})
	Pop() interface{}
	NewQueue() NewQueueFunc
}

type NewQueueFunc func() Queue
