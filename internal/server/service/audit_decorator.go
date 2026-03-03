package service

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/audit"
	"github.com/mrPTqp/metrics/internal/contextkey"
	"go.uber.org/zap"
)

// Декоратор для аудита
type AuditService struct {
	writer   MetricWriter
	eventBus *audit.EventBus
}

// Возвращает новый экземпляр AuditService
func NewAuditService(writer MetricWriter, bus *audit.EventBus) *AuditService {
	return &AuditService{
		writer:   writer,
		eventBus: bus,
	}
}

// Прокидывает сохранение gauge метрики в сервис и отправляет в шину событие аудита
func (a *AuditService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	if err := a.writer.SaveGaugeMetric(ctx, mName, mValue); err != nil {
		return err
	}
	a.logEvent(ctx, []string{mName}, contextkey.LoggerFromContext(ctx))

	return nil
}

// Прокидывает сохранение counter метрики в сервис и отправляет в шину событие аудита
func (a *AuditService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	if err := a.writer.SaveCounterMetric(ctx, mName, mValue); err != nil {
		return err
	}
	a.logEvent(ctx, []string{mName}, contextkey.LoggerFromContext(ctx))
	return nil
}

// Прокидывает сохранение gauge и counter метрик в сервис и отправляет в шину событие аудита
func (a *AuditService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	if err := a.writer.SaveAllMetrics(ctx, gauges, counters); err != nil {
		return err
	}
	var names []string
	for name := range gauges {
		names = append(names, name)
	}
	for name := range counters {
		names = append(names, name)
	}
	a.logEvent(ctx, names, contextkey.LoggerFromContext(ctx))
	return nil
}

func (a *AuditService) logEvent(ctx context.Context, mNames []string, logger *zap.Logger) {
	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   mNames,
		IPAddress: contextkey.GetClientIP(ctx),
	}
	if err := a.eventBus.Publish(event); err != nil {
		logger.Warn("error publish audit event", zap.Any("event", err))
	}
}
