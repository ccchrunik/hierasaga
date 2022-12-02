package service

import (
	"atm/ds"
	"fmt"
	"sync"
)

type Report struct {
	table map[string]*ds.MutexQueue
	mu    sync.Mutex
}

func NewReport() *Report {
	return &Report{
		table: map[string]*ds.MutexQueue{},
	}
}

func (r *Report) Add(serviceName string, round int, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.table[serviceName].Push(fmt.Sprintf("[%d] %s: %v", round, serviceName, v))
}

func (r *Report) Print(serviceName string) {
	entry := r.table[serviceName]
	for i := 0; i < entry.Len(); i++ {
		fmt.Println(entry.Pop())
	}
}

func (r *Report) PrintAll() {
	for serviceName := range r.table {
		r.Print(serviceName)
	}
}
