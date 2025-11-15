package service

import (
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/repository"
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

func (ms *BaseMetricService) SaveGaugeMetric(mName string, mValue *float64) error {
	err := ms.repo.SaveGauge(mName, mValue)
	if err != nil {
		return err
	}
	ms.logger.Infof("Saved gauge: %s = %f", mName, *mValue)
	return nil
}

func (ms *BaseMetricService) SaveCounterMetric(mName string, mValue *int64) error {
	err := ms.repo.SaveCounter(mName, mValue)
	if err != nil {
		return err
	}
	ms.logger.Infof("Saved counter: %s = %d", mName, *mValue)
	return nil
}

func (ms *BaseMetricService) GetGaugeMetric(mName string) (float64, error) {
	mValue, err := ms.repo.GetGauge(mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Retrieved gauge: %s = %f", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) GetCounterMetric(mName string) (int64, error) {
	mValue, err := ms.repo.GetCounter(mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Retrieved counter: %s = %d", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) ListAllMetrics() (map[string]float64, map[string]int64) {
	gauges, err := ms.repo.ListGauges()
	if err != nil {
		ms.logger.Errorf("Failed to list gauges: %v", err)
		gauges = make(map[string]float64)
	}

	counters, err := ms.repo.ListCounters()
	if err != nil {
		ms.logger.Errorf("Failed to list counters: %v", err)
		counters = make(map[string]int64)
	}

	return gauges, counters
}

func (ms *BaseMetricService) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	return ms.repo.SaveAllMetrics(gauges, counters)
}
