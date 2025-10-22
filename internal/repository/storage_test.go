package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_ListMethods(t *testing.T) {
	s := NewMemStorage()
	_ = s.AddGauge("cpu", 2.5)
	_ = s.AddCounter("hits", 3)

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
