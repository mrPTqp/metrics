package handler

import (
	"github.com/mrPTqp/metrics/internal/service"
	"go.uber.org/zap"
)

type MetricHandler struct {
	service service.MetricsService
	logger  *zap.SugaredLogger
}

func NewMetricHandler(service service.MetricsService, logger *zap.SugaredLogger) *MetricHandler {
	return &MetricHandler{
		service: service,
		logger:  logger,
	}
}
