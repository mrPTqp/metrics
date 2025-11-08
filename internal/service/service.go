package service

import (
	"github.com/mrPTqp/metrics/internal/repository"
	"go.uber.org/zap"
)

type MetricsService interface {
	SaveGaugeMetric(mName string, mValue *float64) error
	SaveCounterMetric(mName string, mValue *int64) error
	GetGaugeMetric(mName string) (float64, error)
	GetCounterMetric(mName string) (int64, error)
	ListAllMetrics() (map[string]float64, map[string]int64)
}

type BaseMetricService struct {
	mr         repository.MetricRepository
	syncBackup bool
	logger     *zap.SugaredLogger
}

func NewMetricsService(mr repository.MetricRepository, logger *zap.SugaredLogger, syncBackup bool) *BaseMetricService {
	return &BaseMetricService{
		mr:     mr,
		syncBackup: syncBackup,
		logger: logger,
	}
}

func (ms *BaseMetricService) SaveGaugeMetric(mName string, mValue *float64) error {
	err := ms.mr.AddGauge(mName, mValue)
	if err != nil {
		return err
	}
	ms.logger.Infof("Successfully saved metric - Type: gauge, Name: %s, Value: %f\n", mName, mValue)
	return nil
}

func (ms *BaseMetricService) SaveCounterMetric(mName string, mValue *int64) error {
	err := ms.mr.AddCounter(mName, mValue)
	if err != nil {
		return err
	}
	ms.logger.Infof("Successfully saved metric - Type: counter, Name: %s, Value: %d\n", mName, mValue)
	return nil
}

func (ms *BaseMetricService) GetGaugeMetric(mName string) (float64, error) {
	mValue, err := ms.mr.GetGauge(mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Successfully return metric - Type: gauge, Name: %s, Value: %f\n", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) GetCounterMetric(mName string) (int64, error) {
	mValue, err := ms.mr.GetCounter(mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Successfully return metric - Type: counter, Name: %s, Value: %d\n", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) ListAllMetrics() (map[string]float64, map[string]int64) {
	return ms.mr.ListGauges(), ms.mr.ListCounters()
}
