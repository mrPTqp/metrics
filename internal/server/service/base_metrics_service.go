package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/repository"
)

// Базовый сервис для обработки метрик
type BaseMetricService struct {
	repo   repository.MetricRepository
	logger *zap.Logger
}

// Возвращает новый экземпляр BaseMetricService
func NewMetricsService(repo repository.MetricRepository, logger *zap.Logger) *BaseMetricService {
	return &BaseMetricService{
		repo:   repo,
		logger: logger,
	}
}

// Прокси
func (ms *BaseMetricService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	err := ms.repo.SaveGauge(ctx, mName, mValue)
	if err != nil {
		return err
	}
	log := contextkey.LoggerFromContext(ctx)
	log.Debug("saved gauge", zap.String("name", mName), zap.Float64("value", *mValue))
	return nil
}

// Прокси
func (ms *BaseMetricService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	err := ms.repo.SaveCounter(ctx, mName, mValue)
	if err != nil {
		return err
	}
	log := contextkey.LoggerFromContext(ctx)
	log.Debug("saved counter", zap.String("name", mName), zap.Int64("value", *mValue))
	return nil
}

// Прокси
func (ms *BaseMetricService) GetGaugeMetric(ctx context.Context, mName string) (float64, error) {
	mValue, err := ms.repo.GetGauge(ctx, mName)
	if err != nil {
		return 0, err
	}
	log := contextkey.LoggerFromContext(ctx)
	log.Debug("retrieved gauge", zap.String("name", mName), zap.Float64("value", mValue))
	return mValue, nil
}

// Прокси
func (ms *BaseMetricService) GetCounterMetric(ctx context.Context, mName string) (int64, error) {
	mValue, err := ms.repo.GetCounter(ctx, mName)
	if err != nil {
		return 0, err
	}
	log := contextkey.LoggerFromContext(ctx)
	log.Debug("retrieved counter", zap.String("name", mName), zap.Int64("value", mValue))
	return mValue, nil
}

// Прокси
func (ms *BaseMetricService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	gauges, err := ms.repo.ListGauges(ctx)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("failed to list gauges", zap.Error(err))
		gauges = make(map[string]float64)
	}

	counters, err := ms.repo.ListCounters(ctx)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("failed to list counters", zap.Error(err))
		counters = make(map[string]int64)
	}

	return gauges, counters
}

// Прокси
func (ms *BaseMetricService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	return ms.repo.SaveAllMetrics(ctx, gauges, counters)
}

// Прокси
func (ms *BaseMetricService) Ping(ctx context.Context) bool {
	return ms.repo.CheckStorageAvailability(ctx)
}
