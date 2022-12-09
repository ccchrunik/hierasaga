package service

import (
	"atm/ds"
	"fmt"
	"sync"
)

type Report struct {
	table map[string]ds.Queue
	mu    sync.Mutex
}

func NewReport(srvs map[string]Service) *Report {
	r := Report{
		table: map[string]ds.Queue{},
	}
	for srv := range srvs {
		r.table[srv] = ds.NewMutexArrayQueue()
	}
	return &r
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
