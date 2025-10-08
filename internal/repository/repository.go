package repository

type MetricRepository interface {
	AddGauge(key string, value float64) error
	AddCounter(key string, value int64) error
}