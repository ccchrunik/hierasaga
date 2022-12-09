package simulation

import (
	"atm/service"
	"sync"
)

const (
	DefaultRounds = 1000
	DefaultSeed   = 42
)

type Request struct {
	Msg       service.Message
	Timestamp int
}

func NewInitMessages() []Request {
	return []Request{
		{
			Timestamp: 0,
			Msg: service.Message{
				NextService: service.ServiceGateway,
				Endpoint:    "new_payment",
				Stage:       1,
				Body: map[string]interface{}{
					"OrderID":    "order-1",
					"CustomerID": "customer-123",
				},
			},
		},
		// {
		// 	Timestamp: 10,
		// 	Msg: service.Message{
		// 		ServiceTo: service.ServiceGateway,
		// 		Endpoint:  "new_payment",
		// 		Body: map[string]interface{}{
		// 			"OrderID":    "order-2",
		// 			"CustomerID": "customer-456",
		// 		},
		// 	},
		// },
	}
}

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

	gtw := rs.SysConf.GetService(service.ServiceGateway)
	for _, req := range NewInitMessages() {
		gtw.Send(req.Msg, req.Timestamp)
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

	var wg sync.WaitGroup
	for _, srv := range rs.Sys.Services {
		wg.Add(1)
		go func(srv service.Service) {
			defer wg.Done()
			srv.Receive()
		}(srv)
	}
	wg.Done()

	return nil
}
