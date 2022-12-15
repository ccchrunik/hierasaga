package ds_test

import (
	"atm/ds"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayQueue(t *testing.T) {
	items := []string{
		"apple", "orange", "kiwi",
	}

	aq := ds.NewArrayQueue()

	assert.True(t, aq.IsEmpty())
	assert.Equal(t, 0, aq.Len())

	for i, item := range items {
		aq.Push(item)
		assert.False(t, aq.IsEmpty())
		assert.Equal(t, i+1, aq.Len())
	}

	itemsLen := len(items)
	for i, item := range items {
		v := aq.Pop()
		assert.Equal(t, item, v.(string))
		assert.Equal(t, itemsLen-1-i, aq.Len())
	}
}

func TestPriorityQueue(t *testing.T) {
	items := []*ds.Item{
		ds.NewItem(1, "apple"),
		ds.NewItem(5, "orange"),
		ds.NewItem(4, "kiwi"),
		ds.NewItem(3, "banana"),
		ds.NewItem(7, "lemon"),
	}

	pq := ds.NewPriorityQueue()

	for _, item := range items {
		pq.Push(item)
	}

	assert.Equal(t, pq.Len(), len(items))

	got := pq.Pop().(*ds.Item).Value()
	assert.Equal(t, "apple", got)

	got = pq.Pop().(*ds.Item).Value()
	assert.Equal(t, "banana", got)

	got = pq.Pop().(*ds.Item).Value()
	assert.Equal(t, "kiwi", got)

	got = pq.Pop().(*ds.Item).Value()
	assert.Equal(t, "orange", got)

	assert.Equal(t, pq.Len(), 1)

	got = pq.Pop().(*ds.Item).Value()
	assert.Equal(t, "lemon", got)

	assert.Equal(t, pq.Len(), 0)
	assert.True(t, pq.IsEmpty())
}

func TestTimedPriorityQueue(t *testing.T) {
	items := []*ds.Item{
		ds.NewItem(1, "apple"),
		ds.NewItem(5, "orange"),
		ds.NewItem(4, "kiwi"),
		ds.NewItem(3, "banana"),
		ds.NewItem(8, "lemon"),
	}

	round := 0

	tq := ds.NewTimedPriorityQueue(&round)

	for _, item := range items {
		tq.Push(item)
	}

	assert.Equal(t, tq.Len(), 0)
	assert.True(t, tq.IsEmpty())

	round++
	assert.False(t, tq.IsEmpty())
	assert.Equal(t, tq.Len(), 1)
	got := tq.Pop().(*ds.Item).Value()
	assert.Equal(t, "apple", got)

	round += 5
	assert.False(t, tq.IsEmpty())
	assert.Equal(t, tq.Len(), 3)
	got = tq.Pop().(*ds.Item).Value()
	assert.Equal(t, "banana", got)
	got = tq.Pop().(*ds.Item).Value()
	assert.Equal(t, "kiwi", got)
	got = tq.Pop().(*ds.Item).Value()
	assert.Equal(t, "orange", got)

	round++
	assert.Equal(t, tq.Len(), 0)
	assert.True(t, tq.IsEmpty())

	round++
	assert.False(t, tq.IsEmpty())
	assert.Equal(t, tq.Len(), 1)
	got = tq.Pop().(*ds.Item).Value()
	assert.Equal(t, "lemon", got)

	round++
	assert.Equal(t, tq.Len(), 0)
	assert.True(t, tq.IsEmpty())
}

func TestMutexPriorityQueue(t *testing.T) {
	mq := ds.NewMutexPriorityQueue()

	MAX_THREAD := 100000
	var wg sync.WaitGroup
	for i := 0; i < MAX_THREAD; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mq.Push(ds.NewItem(i, strconv.Itoa(i)))
		}(i)
	}
	wg.Wait()

	assert.Equal(t, mq.Len(), MAX_THREAD)
	for i := 0; i < MAX_THREAD/2; i++ {
		got := mq.Pop().(*ds.Item).Value()
		assert.Equal(t, strconv.Itoa(i), got)
	}
	assert.Equal(t, mq.Len(), MAX_THREAD/2)

	mqDup := ds.NewMutexPriorityQueue()
	mqNew := mq.MoveTo(mqDup)
	assert.Equal(t, mqNew, mqDup)

	for i := MAX_THREAD / 2; i < MAX_THREAD; i++ {
		got := mqNew.Pop().(*ds.Item).Value()
		assert.Equal(t, strconv.Itoa(i), got)
	}
	assert.Equal(t, mq.Len(), 0)
	assert.True(t, mq.IsEmpty())
}
