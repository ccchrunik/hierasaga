package service

import (
	"errors"
)

type ConfigEntry struct {
	FailureType FailureType
}

type SystemConfig struct {
	services map[string]Service
	table    map[string]ConfigEntry
	report   *Report
	round    int
}

func NewSystemConfig(srvs map[string]Service) *SystemConfig {
	return &SystemConfig{
		services: srvs,
		table:    map[string]ConfigEntry{},
		report:   NewReport(),
	}
}

func (sc *SystemConfig) Advance() int {
	sc.round++
	return sc.round
}

func (sc *SystemConfig) Get(srv string) (ConfigEntry, bool) {
	entry, ok := sc.table[srv]
	return entry, ok
}

func (sc *SystemConfig) Set(srv string, value ConfigEntry) {
	sc.table[srv] = value
}

func (sc *SystemConfig) IsFailed(srv string) (bool, error) {
	entry, ok := sc.Get(srv)
	if !ok {
		return false, errors.New("entry not exist")
	}
	return entry.FailureType != FailureNone, nil
}

func (sc *SystemConfig) SetFailure(srv string, failureType FailureType) error {
	entry, ok := sc.Get(srv)
	if !ok {
		return errors.New("entry not exist")
	}
	entry.FailureType = failureType
	sc.Set(srv, entry)
	return nil
}

func (sc *SystemConfig) Log(srv string, msg interface{}) {
	sc.report.Add(srv, sc.round, msg)
}
