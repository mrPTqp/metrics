package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMemStorage_ListMethods(t *testing.T) {
	s := NewMemStorage(zap.NewNop().Sugar())
	cpu := 2.5
	hits := int64(3)
	s.AddGauge("cpu", &cpu)
	s.AddCounter("hits", &hits)

	gauges := s.ListGauges()
	counters := s.ListCounters()

	assert.Equal(t, 2.5, gauges["cpu"])
	assert.EqualValues(t, 3, counters["hits"])

	gauges["cpu"] = 9.9
	counters["hits"] = 999
	g2 := s.ListGauges()
	c2 := s.ListCounters()
	assert.Equal(t, 2.5, g2["cpu"])      // ensure copy
	assert.EqualValues(t, 3, c2["hits"]) // ensure copy
}
