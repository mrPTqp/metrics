package repository

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	logger   *zap.SugaredLogger
	mu       sync.RWMutex
}

func NewMemStorage(logger *zap.SugaredLogger) *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		logger:   logger,
	}
}

func (s *MemStorage) AddGauge(key string, value *float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[key] = *value
	s.logState()
	return nil
}

func (s *MemStorage) AddCounter(key string, value *int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[key] += *value
	s.logState()
	return nil
}

func (s *MemStorage) GetGauge(key string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.gauges[key]; ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("gauge %s not found", key)
	}
}

func (s *MemStorage) GetCounter(key string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.counters[key]; ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("counter %s not found", key)
	}
}

func (s *MemStorage) ListGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyMap := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		copyMap[k] = v
	}
	return copyMap
}

func (s *MemStorage) ListCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyMap := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		copyMap[k] = v
	}
	return copyMap
}

func (s *MemStorage) logState() {
	var log = s.logger
	log.Infoln("Current MemStorage state:")
	log.Infoln("Gauges:")
	for k, v := range s.gauges {
		log.Infof("  %s: %f\n", k, v)
	}
	log.Infoln("Counters:")
	for k, v := range s.counters {
		log.Infof("  %s: %d\n", k, v)
	}
	log.Infoln("-------------------------")
}
