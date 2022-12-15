package simulation_test

import (
	"atm/simulation"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulate(t *testing.T) {
	simulator := simulation.NewRoundSimultor()
	simConf := simulation.NewSimulationConfig()
	simulator.Simulate(*simConf)
	assert.True(t, true)
}
