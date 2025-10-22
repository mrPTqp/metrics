package repository

type MetricRepository interface {
	AddGauge(key string, value float64) error
	AddCounter(key string, value int64) error
	GetGauge(key string) (float64, error)
	GetCounter(key string) (int64, error)
	ListGauges() map[string]float64
	ListCounters() map[string]int64
}
