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
	msr    repository.MetricRepository
	fsr    repository.MetricRepository
	logger *zap.SugaredLogger
}

func NewMetricsService(msr repository.MetricRepository, fsr repository.MetricRepository, logger *zap.SugaredLogger) *BaseMetricService {
	return &BaseMetricService{
		msr:    msr,
		fsr:    fsr,
		logger: logger,
	}
}

func (ms *BaseMetricService) SaveGaugeMetric(mName string, mValue *float64) error {
	err := ms.msr.SaveGauge(mName, mValue)
	if err != nil {
		return err
	}
	err1 := ms.fsr.SaveGauge(mName, mValue)
	if err1 != nil {
		return err1
	}

	ms.logger.Infof("Successfully saved metric - Type: gauge, Name: %s, Value: %f\n", mName, mValue)
	return nil
}

func (ms *BaseMetricService) SaveCounterMetric(mName string, mValue *int64) error {
	err := ms.msr.SaveCounter(mName, mValue)
	if err != nil {
		return err
	}
	err1 := ms.fsr.SaveCounter(mName, mValue)
	if err1 != nil {
		return err1
	}

	ms.logger.Infof("Successfully saved metric - Type: counter, Name: %s, Value: %d\n", mName, mValue)
	return nil
}

func (ms *BaseMetricService) GetGaugeMetric(mName string) (float64, error) {
	mValue, err := ms.msr.GetGauge(mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Successfully return metric - Type: gauge, Name: %s, Value: %f\n", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) GetCounterMetric(mName string) (int64, error) {
	mValue, err := ms.msr.GetCounter(mName)
	if err != nil {
		return 0, err
	}
	ms.logger.Infof("Successfully return metric - Type: counter, Name: %s, Value: %d\n", mName, mValue)
	return mValue, nil
}

func (ms *BaseMetricService) ListAllMetrics() (map[string]float64, map[string]int64) {
	gauges, err := ms.msr.ListGauges()
	if err != nil {
		ms.logger.Errorf("Failed to retrieve gauge metrics: %v", err)
		gauges = make(map[string]float64)
	}

	counters, err := ms.msr.ListCounters()
	if err != nil {
		ms.logger.Errorf("Failed to retrieve counter metrics: %v", err)
		counters = make(map[string]int64)
	}

	return gauges, counters
}

func (ms *BaseMetricService) SaveAllMetricsToFile(gauges map[string]float64, counters map[string]int64) error {
	err := ms.fsr.SaveAllMetrics(gauges, counters)
	if err != nil {
		return err
	}

	return nil
}