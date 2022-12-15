package simulation

import (
	"atm/service"
)

const (
	DefaultRounds = 20
	DefaultSeed   = 42
)

type Request struct {
	Req       service.Request
	Timestamp int
}

func NewInitRequest() []Request {
	return []Request{
		{
			Timestamp: 0,
			Req: service.Request{
				Service:  service.ServicePayment,
				Endpoint: "payment_control",
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

func NewSimulationConfig() *SimulationConfig {
	return &SimulationConfig{}
}

type Simulator interface {
	Simulate(SimulationConfig) error
}

type RoundSimulator struct {
	Sys     *service.System
	SimConf SimulationConfig
}

func NewRoundSimultor() *RoundSimulator {
	return &RoundSimulator{}
}

func (rs *RoundSimulator) Simulate(simConf SimulationConfig) error {
	sys := service.NewSystem()

	if err := rs.init(sys, simConf); err != nil {
		return err
	}

	gtw := rs.Sys.Gateway
	for _, req := range NewInitRequest() {
		gtw.Send(req.Req, req.Timestamp)
	}

	for r := 0; r < rs.SimConf.Rounds; r++ {
		if err := rs.run(); err != nil {
			return err
		}
	}

	return nil
}

func (rs *RoundSimulator) init(sys *service.System, simConf SimulationConfig) error {
	// TODO:
	// if err := simConf.Pattern.Init(); err != nil {
	// 	return err
	// }
	if simConf.Rounds == 0 {
		simConf.Rounds = DefaultRounds
	}
	if simConf.Seed == 0 {
		simConf.Seed = 42
	}

	rs.Sys = sys
	rs.SimConf = simConf
	return nil
}

func (rs *RoundSimulator) run() error {
	// round := rs.Sys.Advance()

	// fmt.Printf("round: %d\n", rs.Sys.Round())
	// for _, srv := range rs.Sys.Services {
	// 	srvName := srv.Name()
	// 	failureType, _ := rs.SimConf.Pattern.Get(srvName, round)
	// 	rs.Sys.SetFailure(srvName, failureType)
	// }

	rs.Sys.Gateway.Receive()

	// var wg sync.WaitGroup
	// for _, srv := range rs.Sys.Services {
	// 	wg.Add(1)
	// 	go func(srv service.Service) {
	// 		defer wg.Done()
	// 		srv.Receive()
	// 	}(srv)
	// }
	// wg.Done()
	for _, srv := range rs.Sys.Services {
		srv.Receive()
	}

	rs.Sys.PrintResult()
	rs.Sys.Advance()

	return nil
}
