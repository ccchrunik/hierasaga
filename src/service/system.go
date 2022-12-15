package service

type StatusEntry struct {
	FailureType FailureType
}

type SystemConfig struct {
	status map[string]StatusEntry
	report *Report
	round  int
}

func NewSystemConfig(srvs []string) *SystemConfig {
	sc := SystemConfig{
		status: map[string]StatusEntry{},
		report: NewReport(),
		round:  0,
	}

	for _, srv := range srvs {
		sc.status[srv] = StatusEntry{
			FailureType: FailureNone,
		}
	}
	return &sc
}

type System struct {
	Gateway    *RoundGateway
	EventQueue *EventQueue
	Services   map[string]Service
	Cfg        *SystemConfig
}

func NewSystem() *System {
	srvs := []string{
		ServiceGateway,
		ServiceEventQueue,
		ServiceTxManager,
		ServicePayment,
		ServiceOrder,
		ServiceCustomer,
		ServiceShipping,
		ServiceNotification,
	}
	sys := System{
		Services: map[string]Service{},
		Cfg:      NewSystemConfig(srvs),
	}

	sys.Gateway = NewRoundGateway(&sys)
	sys.EventQueue = NewEventQueue(&sys)

	sys.Services[ServiceTxManager] = NewTxManager(&sys)
	sys.Services[ServicePayment] = NewPaymentService(&sys)
	sys.Services[ServiceOrder] = NewOrderService(&sys)
	sys.Services[ServiceShipping] = NewShippingService(&sys)
	sys.Services[ServiceCustomer] = NewCustomerService(&sys)
	sys.Services[ServiceNotification] = NewNotificationService(&sys)

	return &sys
}

func (sys *System) Advance() int {
	sys.Cfg.round++
	return sys.Cfg.round
}

func (sys *System) Round() int {
	return sys.Cfg.round
}

func (sys *System) GetService(srv string) Service {
	return sys.Services[srv]
}

func (sys *System) GetStatus(srv string) StatusEntry {
	return sys.Cfg.status[srv]
}

func (sys *System) SetStatus(srv string, entry StatusEntry) {
	sys.Cfg.status[srv] = entry
}

func (sys *System) IsFailed(srv string) (bool, error) {
	entry := sys.GetStatus(srv)
	return entry.FailureType != FailureNone, nil
}

func (sys *System) SetFailure(srv string, failureType FailureType) error {
	entry := sys.GetStatus(srv)
	entry.FailureType = failureType
	sys.SetStatus(srv, entry)
	return nil
}

func (sys *System) Log(srv string, msg interface{}) {
	sys.Cfg.report.Add(srv, sys.Cfg.round, msg)
}

func (sys *System) PrintResult() {
	// sys.Cfg.report.SortAll()
	// sys.Cfg.report.PrintAll()
	// sys.Cfg.report.ClearAll()
}
