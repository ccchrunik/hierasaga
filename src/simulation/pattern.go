package simulation

import "atm/service"

var BasicPattern DefinedIntervalPattern

func init() {
	BasicPattern = NewDefinedIntervalPattern()
	BasicPattern.IntervalMap = map[string][]Interval{
		service.ServiceGateway: {
			{
				Start: 1,
				End:   2,
			},
			{
				Start: 3,
				End:   4,
			},
			{
				Start: 7,
				End:   8,
			},
		},
		service.ServiceCustomer: {
			{
				Start: 1,
				End:   5,
			},
			{
				Start: 3,
				End:   7,
			},
			{
				Start: 8,
				End:   9,
			},
			{
				Start: 12,
				End:   15,
			},
		},
	}
}

type FailurePattern interface {
	Init() error
	Get(string, int) (service.FailureType, bool)
}

type DefinedIntervalPattern struct {
	IntervalMap map[string][]Interval
	ProgressMap map[string]int
}

func NewDefinedIntervalPattern() DefinedIntervalPattern {
	return DefinedIntervalPattern{
		IntervalMap: map[string][]Interval{},
		ProgressMap: map[string]int{},
	}
}

func (p *DefinedIntervalPattern) Init() error {
	if err := p.sort(); err != nil {
		return err
	}
	if err := p.merge(); err != nil {
		return err
	}

	for service := range p.IntervalMap {
		p.ProgressMap[service] = 0
	}
	return nil
}

func (p *DefinedIntervalPattern) Get(srv string, round int) (service.FailureType, bool) {
	if !p.hasNext(srv) {
		return service.FailureNone, false
	}
	idx := p.ProgressMap[srv]
	interval := p.IntervalMap[srv][idx]
	if round >= interval.Start && round <= interval.End {
		return interval.FailureType, true
	} else if round < interval.Start {
		return service.FailureNone, false
	}
	p.advance(srv)
	return service.FailureNone, false
}

func (p *DefinedIntervalPattern) hasNext(srv string) bool {
	return p.ProgressMap[srv] < len(p.IntervalMap[srv])
}

func (p *DefinedIntervalPattern) advance(srv string) (Interval, error) {
	idx := p.ProgressMap[srv]
	p.ProgressMap[srv]++
	return p.IntervalMap[srv][idx], nil
}

func (p *DefinedIntervalPattern) sort() error {
	start := func(p1, p2 *Interval) bool {
		return p1.Start < p2.Start
	}

	for _, intervals := range p.IntervalMap {
		By(start).Sort(intervals)
	}

	return nil
}

func (p *DefinedIntervalPattern) merge() error {
	for service, intervals := range p.IntervalMap {
		newIntervals := []Interval{}
		if len(intervals) == 0 {
			continue
		}

		interval := intervals[0]
		for i := 1; i < len(intervals); i++ {
			nextInterval := intervals[i]
			if nextInterval.Start >= interval.Start && nextInterval.Start <= interval.End+1 {
				interval.End = nextInterval.End
			} else {
				newIntervals = append(newIntervals, interval)
				interval = nextInterval
			}
		}
		p.IntervalMap[service] = append(newIntervals, interval)
	}

	return nil
}
