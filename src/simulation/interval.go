package simulation

import (
	"atm/service"
	"sort"
)

type Interval struct {
	Start       int
	End         int
	Status      service.Status
	FailureType service.FailureType
}

// https://pkg.go.dev/sort
type By func(p1, p2 *Interval) bool

func (by By) Sort(intervals []Interval) {
	is := &intervalSorter{
		intervals: intervals,
		by:        by,
	}
	sort.Sort(is)
}

type intervalSorter struct {
	intervals []Interval
	by        func(p1, p2 *Interval) bool // Closure used in the Less method.
}

func (s *intervalSorter) Len() int {
	return len(s.intervals)
}

func (s *intervalSorter) Swap(i, j int) {
	s.intervals[i], s.intervals[j] = s.intervals[j], s.intervals[i]
}

func (s *intervalSorter) Less(i, j int) bool {
	return s.by(&s.intervals[i], &s.intervals[j])
}
