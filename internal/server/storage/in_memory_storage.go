package storage

import (
	"context"
	"errors"
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

func (s *MemStorage) SaveGauge(_ context.Context, name string, value *float64) error {
	if name == "" || value == nil {
		return errors.New("invalid gauge metric: empty name or nil value")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = *value
	return nil
}

func (s *MemStorage) SaveCounter(_ context.Context, name string, value *int64) error {
	if name == "" || value == nil {
		return errors.New("invalid counter metric: empty name or nil value")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += *value
	return nil
}

func (s *MemStorage) GetGauge(_ context.Context, name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.gauges[name]; ok {
		return val, nil
	}
	return 0, ErrMetricNotFound
}

func (s *MemStorage) GetCounter(_ context.Context, name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.counters[name]; ok {
		return val, nil
	}
	return 0, ErrMetricNotFound
}

func (s *MemStorage) ListGauges(_ context.Context) (map[string]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.gauges), nil
}

func (s *MemStorage) ListCounters(_ context.Context) (map[string]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.counters), nil
}

func (s *MemStorage) SaveAllMetrics(_ context.Context, gauges map[string]float64, counters map[string]int64) error {
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

	return nil
}

func (s *MemStorage) CheckStorageAvailability(_ context.Context) bool {
	return true
}

func (s *MemStorage) Close() error {
	return nil
}
