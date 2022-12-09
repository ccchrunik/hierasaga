package ds

import "container/heap"

// https://pkg.go.dev/container/heap

// An Item is something we manage in a priority queue.
type Item struct {
	priority int         // The priority of the item in the queue.
	value    interface{} // The value of the item; arbitrary.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

func NewItem(priority int, value interface{}) *Item {
	return &Item{
		priority: priority,
		value:    value,
	}
}

func (item *Item) Value() interface{} {
	return item.value
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueueInner []*Item

func (pq *PriorityQueueInner) IsEmpty() bool {
	return pq.Len() == 0
}

func (pq PriorityQueueInner) Len() int { return len(pq) }

func (pq PriorityQueueInner) Less(i, j int) bool {
	// the item with the least timestamp has the highest priority
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueueInner) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueueInner) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueueInner) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
// func (pq *PriorityQueue) update(item *Item, value interface{}, timestamp int) {
// 	item.value = value
// 	item.timestamp = timestamp
// 	heap.Fix(pq, item.index)
// }

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueueInner{}
	heap.Init(pq)
	return &PriorityQueue{
		queue: pq,
	}
}

type PriorityQueue struct {
	queue *PriorityQueueInner
}

func (pqw *PriorityQueue) NewQueue() NewQueueFunc {
	return func() Queue {
		return NewPriorityQueue()
	}
}

func (pqw *PriorityQueue) IsEmpty() bool {
	return pqw.queue.IsEmpty()
}

func (pqw *PriorityQueue) Len() int {
	return pqw.queue.Len()
}

func (pqw *PriorityQueue) Push(v interface{}) {
	heap.Push(pqw.queue, v)
}

func (pqw *PriorityQueue) Pop() interface{} {
	return heap.Pop(pqw.queue)
}
