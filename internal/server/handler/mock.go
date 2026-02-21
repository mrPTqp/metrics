package handler

import (
	"context"

	"github.com/mrPTqp/metrics/internal/server/service"
)

type mockMetricsService struct {
	saveGaugeFunc      func(ctx context.Context, name string, value *float64) error
	saveCounterFunc    func(ctx context.Context, name string, delta *int64) error
	getGaugeFunc       func(ctx context.Context, name string) (float64, error)
	getCounterFunc     func(ctx context.Context, name string) (int64, error)
	listAllFunc        func(ctx context.Context) (map[string]float64, map[string]int64)
	pingFunc           func(ctx context.Context) bool
	saveAllMetricsFunc func(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
}

var _ service.MetricsService = (*mockMetricsService)(nil)

func (m *mockMetricsService) SaveGaugeMetric(ctx context.Context, name string, value *float64) error {
	return m.saveGaugeFunc(ctx, name, value)
}

func (m *mockMetricsService) SaveCounterMetric(ctx context.Context, name string, delta *int64) error {
	return m.saveCounterFunc(ctx, name, delta)
}

func (m *mockMetricsService) GetGaugeMetric(ctx context.Context, name string) (float64, error) {
	return m.getGaugeFunc(ctx, name)
}

func (m *mockMetricsService) GetCounterMetric(ctx context.Context, name string) (int64, error) {
	return m.getCounterFunc(ctx, name)
}

func (m *mockMetricsService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return map[string]float64{}, map[string]int64{}
}

func (m *mockMetricsService) Ping(ctx context.Context) bool {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return true
}

func (m *mockMetricsService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	if m.saveAllMetricsFunc != nil {
		return m.saveAllMetricsFunc(ctx, gauges, counters)
	}
	return nil
}
