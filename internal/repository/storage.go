package repository

import "fmt"

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) AddGauge(key string, value float64) error {
	s.gauges[key] = value
	s.logState()
	return nil
}

func (s *MemStorage) AddCounter(key string, value int64) error {
	s.counters[key] += value
	s.logState()
	return nil
}

func (s *MemStorage) GetGauge(key string) (float64, error) {
	if val, ok := s.gauges[key]; ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("gauge %s not found", key)
	}
}

func (s *MemStorage) GetCounter(key string) (int64, error) {
	if val, ok := s.counters[key]; ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("counter %s not found", key)
	}
}

func (s *MemStorage) ListGauges() map[string]float64 {
	copyMap := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		copyMap[k] = v
	}
	return copyMap
}

func (s *MemStorage) ListCounters() map[string]int64 {
	copyMap := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		copyMap[k] = v
	}
	return copyMap
}

func (s *MemStorage) logState() {
	fmt.Println("Current MemStorage state:")
	fmt.Println("Gauges:")
	for k, v := range s.gauges {
		fmt.Printf("  %s: %f\n", k, v)
	}
	fmt.Println("Counters:")
	for k, v := range s.counters {
		fmt.Printf("  %s: %d\n", k, v)
	}
	fmt.Println("-------------------------")
}