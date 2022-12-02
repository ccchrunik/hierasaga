package simulation

import (
	"atm/service"
	"sync"
)

const (
	DefaultRounds = 1_000_000
	DefaultSeed   = 42
)

type SimulationConfig struct {
	Pattern FailurePattern `json:"pattern"`
	Rounds  int            `json:"rounds"`
	Seed    int            `json:"seed"`
}

type Simulator interface {
	Simulate(SimulationConfig) error
}

type RoundSimulator struct {
	Sys     *service.System
	SysConf *service.SystemConfig
	SimConf SimulationConfig
}

func (rs *RoundSimulator) Simulate(simConf SimulationConfig) error {
	sys := service.NewSystem()

	if err := rs.init(sys, simConf); err != nil {
		return err
	}

	for r := 0; r < rs.SimConf.Rounds; r++ {
		if err := rs.run(); err != nil {
			return err
		}
	}

	return nil
}

func (rs *RoundSimulator) init(sys *service.System, simConf SimulationConfig) error {
	if err := simConf.Pattern.Init(); err != nil {
		return err
	}
	if simConf.Rounds == 0 {
		simConf.Rounds = DefaultRounds
	}
	if simConf.Seed == 0 {
		simConf.Seed = 42
	}

	rs.Sys = sys
	rs.SimConf = simConf

	cfg := service.NewSystemConfig(sys.Services)
	rs.SysConf = cfg
	return nil
}

func (rs *RoundSimulator) run() error {
	round := rs.SysConf.Advance()

	for _, srv := range rs.Sys.Services {
		srvName := srv.Name()
		failureType, _ := rs.SimConf.Pattern.Get(srvName, round)
		rs.SysConf.SetFailure(srvName, failureType)
	}

	var wgSetup sync.WaitGroup
	for _, srv := range rs.Sys.Services {
		wgSetup.Add(1)
		go func(srv service.Service) {
			defer wgSetup.Done()
			if err := srv.Setup(); err != nil {
				rs.SysConf.Log(srv.Name(), err)
			}
		}(srv)
	}
	wgSetup.Done()

	var wgExec sync.WaitGroup
	for _, srv := range rs.Sys.Services {
		wgExec.Add(1)
		go func(srv service.Service) {
			defer wgExec.Done()
			if err := srv.Execute(); err != nil {
				rs.SysConf.Log(srv.Name(), err)
			}
		}(srv)
	}
	wgExec.Done()

	return nil
}
