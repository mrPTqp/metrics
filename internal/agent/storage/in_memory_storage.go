package storage

import (
	"maps"
	"sync"
)

type MemStorage struct {
	gauges           map[string]float64
	counters         map[string]int64
	additionalGauges map[string]float64
	mu               sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:           make(map[string]float64),
		counters:         make(map[string]int64),
		additionalGauges: make(map[string]float64),
	}
}

func (s *MemStorage) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if gauges == nil {
		s.gauges = make(map[string]float64)
	} else {
		s.gauges = maps.Clone(gauges)
	}

	if counters == nil {
		s.counters = make(map[string]int64)
	} else {
		s.counters = maps.Clone(counters)
	}
}

func (s *MemStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return maps.Clone(s.gauges), maps.Clone(s.counters)
}

func (s *MemStorage) SaveAdditionalGaugeMetrics(additionalGauges map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if additionalGauges == nil {
		s.additionalGauges = make(map[string]float64)
	} else {
		s.additionalGauges = maps.Clone(additionalGauges)
	}
}

func (s *MemStorage) GetAdditionalGaugeMetrics() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return maps.Clone(s.additionalGauges)
}
