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