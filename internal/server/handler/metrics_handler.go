package handler

import (
	"github.com/mrPTqp/metrics/internal/server/service"
	"go.uber.org/zap"
)

// Обработчик HTTP запросов
type MetricHandler struct {
	service service.MetricsService
	logger  *zap.Logger
}

// Возвращает новый экземпляр MetricHandler
func NewMetricHandler(service service.MetricsService, logger *zap.Logger) *MetricHandler {
	return &MetricHandler{
		service: service,
		logger:  logger,
	}
}
