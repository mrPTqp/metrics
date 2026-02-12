package service

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/audit"
	"github.com/mrPTqp/metrics/internal/contextkey"
)

// Декоратор для аудита
type AuditService struct {
	service  MetricsService
	eventBus *audit.EventBus
}

// Возвращает новый экземпляр AuditService
func NewAuditService(service MetricsService, bus *audit.EventBus) *AuditService {
	return &AuditService{
		service:  service,
		eventBus: bus,
	}
}

// Прокидывает сохранение gauge метрики в сервис и отправляет в шину событие аудита
func (a *AuditService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	if err := a.service.SaveGaugeMetric(ctx, mName, mValue); err != nil {
		return err
	}
	a.logEvent(ctx, []string{mName})

	return nil
}

// Прокидывает сохранение counter метрики в сервис и отправляет в шину событие аудита
func (a *AuditService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	if err := a.service.SaveCounterMetric(ctx, mName, mValue); err != nil {
		return err
	}
	a.logEvent(ctx, []string{mName})
	return nil
}

// Прокидывает сохранение gauge и counter метрик в сервис и отправляет в шину событие аудита
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

// Прокси
func (a *AuditService) GetGaugeMetric(ctx context.Context, name string) (float64, error) {
	return a.service.GetGaugeMetric(ctx, name)
}

// Прокси
func (a *AuditService) GetCounterMetric(ctx context.Context, name string) (int64, error) {
	return a.service.GetCounterMetric(ctx, name)
}

// Прокси
func (a *AuditService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	return a.service.ListAllMetrics(ctx)
}

// Прокси
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
