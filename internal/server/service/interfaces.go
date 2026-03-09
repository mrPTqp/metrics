package service

import "context"

type MetricWriter interface {
    SaveGaugeMetric(ctx context.Context, name string, value *float64) error
    SaveCounterMetric(ctx context.Context, name string, value *int64) error
    SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
}

type MetricReader interface {
    GetGaugeMetric(ctx context.Context, name string) (float64, error)
    GetCounterMetric(ctx context.Context, name string) (int64, error)
}

type MetricsLister interface {
    ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64)
}

type Pinger interface {
    Ping(ctx context.Context) bool
}

// PingerFunc is an adapter to allow the use of ordinary functions as Pinger.
type PingerFunc func(ctx context.Context) bool

func (f PingerFunc) Ping(ctx context.Context) bool {
	return f(ctx)
}