package service

import (
	"fmt"
	"strconv"

	"github.com/mrPTqp/metrics/internal/repository"
)

type MetricsService struct {
	mr repository.MetricRepository
}

func NewMetricsService(mr repository.MetricRepository) *MetricsService {
	return &MetricsService{
		mr: mr,
	}
}

func (ms *MetricsService) ProcessMetric(metricType, name, value string) error {
	switch metricType {
	case "gauge":
		specifiedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}

		err = ms.mr.AddGauge(name, specifiedValue)
		if err != nil {
            return err
        }
	case "counter":
		specifiedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		err = ms.mr.AddCounter(name, specifiedValue)
        if err != nil {
            return err
        }
	default:
        return fmt.Errorf("[ERROR] Invalid metric type")
	}

	fmt.Printf("Successfully saved metric - Type: %s, Name: %s, Value: %s\n", metricType, name, value)
	return nil
}
