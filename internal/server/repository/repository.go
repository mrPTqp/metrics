package repository

import "context"

// Интерфейс репозитория метрик
type MetricRepository interface {
	SaveGauge(ctx context.Context, name string, value *float64) error
	SaveCounter(ctx context.Context, name string, value *int64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	ListGauges(ctx context.Context) (map[string]float64, error)
	ListCounters(ctx context.Context) (map[string]int64, error)
	SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
	CheckStorageAvailability(ctx context.Context) bool
	Close() error
}
