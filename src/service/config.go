package service

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
	sc := SystemConfig{
		services: srvs,
		table:    map[string]ConfigEntry{},
		report:   NewReport(srvs),
	}

	for srv := range srvs {
		sc.table[srv] = ConfigEntry{
			FailureType: FailureNone,
		}
	}
	return &sc
}

func (sc *SystemConfig) Advance() int {
	sc.round++
	return sc.round
}

func (sc *SystemConfig) Round() int {
	return sc.round
}

func (sc *SystemConfig) GetService(srv string) Service {
	return sc.services[srv]
}

func (sc *SystemConfig) GetStatus(srv string) ConfigEntry {
	return sc.table[srv]
}

func (sc *SystemConfig) SetStatus(srv string, value ConfigEntry) {
	sc.table[srv] = value
}

func (sc *SystemConfig) IsFailed(srv string) (bool, error) {
	entry := sc.GetStatus(srv)
	return entry.FailureType != FailureNone, nil
}

func (sc *SystemConfig) SetFailure(srv string, failureType FailureType) error {
	entry := sc.GetStatus(srv)
	entry.FailureType = failureType
	sc.SetStatus(srv, entry)
	return nil
}

func (sc *SystemConfig) Log(srv string, msg interface{}) {
	sc.report.Add(srv, sc.round, msg)
}
