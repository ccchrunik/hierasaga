package service

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
)

type Report struct {
	table map[string][]string
	w     io.Writer
	mu    sync.Mutex
}

func NewReport() *Report {
	r := Report{
		table: map[string][]string{},
		w:     os.Stdout,
	}
	return &r
}

func (r *Report) SetWriter(w io.Writer) {
	r.w = w
}

func (r *Report) Add(srvName string, round int, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := v.(Event)
	s := fmt.Sprintf("[%06d] (%d) TxID: {%s} %s -> [%s/%s/%d]",
		round,
		e.CurrentRetryTime,
		e.TxID,
		e.From,
		e.To,
		e.Endpoint,
		e.Stage)
	// r.table[srvName] = append(r.table[srvName], s)
	fmt.Println(s)
}

func (r *Report) Clear(srvName string) {
	r.table[srvName] = []string{}
}

func (r *Report) Sort(srvName string) {
	sort.Slice(r.table[srvName], func(i, j int) bool {
		return r.table[srvName][i] < r.table[srvName][j]
	})
}

func (r *Report) SortAll() {
	for srvName := range r.table {
		r.Sort(srvName)
	}
}

func (r *Report) ClearAll() {
	for srvName := range r.table {
		r.Clear(srvName)
	}
}

func (r *Report) Print(serviceName string) {
	for s := range r.table {
		fmt.Fprintln(r.w, s)
	}
}

func (r *Report) PrintAll() {
	for serviceName := range r.table {
		r.Print(serviceName)
	}
}
