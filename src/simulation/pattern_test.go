package simulation_test

import (
	"atm/service"
	"atm/simulation"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternInit(t *testing.T) {
	pattern := simulation.NewDefinedIntervalPattern()
	pattern.IntervalMap = map[string][]simulation.Interval{
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

	pattern.Init()

	assert.Equal(t, pattern.IntervalMap[service.ServiceGateway], []simulation.Interval{
		{
			Start: 1,
			End:   4,
		},
		{
			Start: 7,
			End:   8,
		},
	})
	assert.Equal(t, pattern.IntervalMap[service.ServiceCustomer], []simulation.Interval{
		{
			Start: 1,
			End:   9,
		},
		{
			Start: 12,
			End:   15,
		},
	})
}

func TestFailurePattern(t *testing.T) {
	pattern := simulation.NewDefinedIntervalPattern()
	pattern.IntervalMap = map[string][]simulation.Interval{
		service.ServiceGateway: {
			{
				Start:       1,
				End:         2,
				FailureType: service.FailureCrash,
			},
			{
				Start:       7,
				End:         9,
				FailureType: service.FailureLinkBroken,
			},
		},
	}

	pattern.Init()

	resultType, isFailed := pattern.Get(service.ServiceGateway, 0)
	assert.False(t, isFailed)
	assert.Equal(t, resultType, service.FailureNone)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 1)
	assert.True(t, isFailed)
	assert.Equal(t, resultType, service.FailureCrash)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 2)
	assert.True(t, isFailed)
	assert.Equal(t, resultType, service.FailureCrash)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 3)
	assert.False(t, isFailed)
	assert.Equal(t, resultType, service.FailureNone)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 6)
	assert.False(t, isFailed)
	assert.Equal(t, resultType, service.FailureNone)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 7)
	assert.True(t, isFailed)
	assert.Equal(t, resultType, service.FailureLinkBroken)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 8)
	assert.True(t, isFailed)
	assert.Equal(t, resultType, service.FailureLinkBroken)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 9)
	assert.True(t, isFailed)
	assert.Equal(t, resultType, service.FailureLinkBroken)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 10)
	assert.False(t, isFailed)
	assert.Equal(t, resultType, service.FailureNone)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 15)
	assert.False(t, isFailed)
	assert.Equal(t, resultType, service.FailureNone)

	resultType, isFailed = pattern.Get(service.ServiceGateway, 20)
	assert.False(t, isFailed)
	assert.Equal(t, resultType, service.FailureNone)
}
