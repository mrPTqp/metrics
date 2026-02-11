package service

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/audit"
	"github.com/mrPTqp/metrics/internal/contextkey"
)

type AuditService struct {
	service  MetricsService
	eventBus *audit.EventBus
}

func NewAuditService(service MetricsService, bus *audit.EventBus) *AuditService {
	return &AuditService{
		service:  service,
		eventBus: bus,
	}
}

func (a *AuditService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	if err := a.service.SaveGaugeMetric(ctx, mName, mValue); err != nil {
		return err
	}
	a.logEvent(ctx, []string{mName})

	return nil
}

func (a *AuditService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	if err := a.service.SaveCounterMetric(ctx, mName, mValue); err != nil {
		return err
	}
	a.logEvent(ctx, []string{mName})
	return nil
}

func (a *AuditService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	if err := a.service.SaveAllMetrics(ctx, gauges, counters); err != nil {
		return err
	}
	var names []string
	for name := range gauges {
		names = append(names, name)
	}
	for name := range counters {
		names = append(names, name)
	}
	a.logEvent(ctx, names)
	return nil
}

func (a *AuditService) GetGaugeMetric(ctx context.Context, name string) (float64, error) {
	return a.service.GetGaugeMetric(ctx, name)
}

func (a *AuditService) GetCounterMetric(ctx context.Context, name string) (int64, error) {
	return a.service.GetCounterMetric(ctx, name)
}

func (a *AuditService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	return a.service.ListAllMetrics(ctx)
}

func (a *AuditService) Ping(ctx context.Context) bool {
	return a.service.Ping(ctx)
}

func (a *AuditService) logEvent(ctx context.Context, mNames []string) {
	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   mNames,
		IPAddress: contextkey.GetClientIP(ctx),
	}
	a.eventBus.Publish(event)
}
