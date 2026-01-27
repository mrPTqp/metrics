package service

import (
	"context"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/server/repository"
)

type BaseMetricService struct {
	repo   repository.MetricRepository
	logger *zap.SugaredLogger
}

func NewMetricsService(repo repository.MetricRepository, logger *zap.SugaredLogger) *BaseMetricService {
	return &BaseMetricService{
		repo:   repo,
		logger: logger,
	}
}

func (ms *BaseMetricService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	err := ms.repo.SaveGauge(ctx, mName, mValue)
	if err != nil {
		return err
	}
	ms.logger.Infof("Saved gauge: %s = %f", mName, *mValue)
	return nil
}

func (ms *BaseMetricService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	err := ms.repo.SaveCounter(ctx, mName, mValue)
	if err != nil {
		return err
	}
	ms.logger.Infof("Saved counter: %s = %d", mName, *mValue)
	return nil
}

func (ms *BaseMetricService) GetGaugeMetric(ctx context.Context, mName string) (float64, error) {
	mValue, err := ms.repo.GetGauge(ctx, mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Retrieved gauge: %s = %f", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) GetCounterMetric(ctx context.Context, mName string) (int64, error) {
	mValue, err := ms.repo.GetCounter(ctx, mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Retrieved counter: %s = %d", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	gauges, err := ms.repo.ListGauges(ctx)
	if err != nil {
		ms.logger.Errorf("Failed to list gauges: %v", err)
		gauges = make(map[string]float64)
	}

	counters, err := ms.repo.ListCounters(ctx)
	if err != nil {
		ms.logger.Errorf("Failed to list counters: %v", err)
		counters = make(map[string]int64)
	}

	return gauges, counters
}

func (ms *BaseMetricService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	return ms.repo.SaveAllMetrics(ctx, gauges, counters)
}

func (ms *BaseMetricService) Ping(ctx context.Context) bool {
	return ms.repo.CheckStorageAvailability(ctx)
}
