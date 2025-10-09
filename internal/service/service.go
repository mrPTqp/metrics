package service

import (
	"fmt"

	"github.com/mrPTqp/metrics/internal/repository"
)

type MetricsService interface {
	ProcessGaugeMetric(mName string, mValue float64) error
	ProcessCounterMetric(mName string, mValue int64) error
}

type BaseMetricService struct {
	mr repository.MetricRepository
}

func NewMetricsService(mr repository.MetricRepository) *BaseMetricService {
	return &BaseMetricService{
		mr: mr,
	}
}

func (ms *BaseMetricService) ProcessGaugeMetric(mName string, mValue float64) error {
	err := ms.mr.AddGauge(mName, mValue)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully saved metric - Type: gauge, Name: %s, Value: %f\n", mName, mValue)
	return nil
}

func (ms *BaseMetricService) ProcessCounterMetric(mName string, mValue int64) error {
	err := ms.mr.AddCounter(mName, mValue)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully saved metric - Type: counter, Name: %s, Value: %d\n", mName, mValue)
	return nil
}
