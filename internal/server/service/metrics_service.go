package service

import "context"

// Интерфейс сервиса метрик
type MetricsService interface {
	SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error
	SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error
	GetGaugeMetric(ctx context.Context, mName string) (float64, error)
	GetCounterMetric(ctx context.Context, mName string) (int64, error)
	ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64)
	SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
	Ping(ctx context.Context) bool
}
