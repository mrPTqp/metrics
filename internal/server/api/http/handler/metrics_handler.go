package handler

import (
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/server/service"
)

// Обработчик HTTP запросов
type MetricHandler struct {
	writer service.MetricWriter
	reader service.MetricReader
	lister service.MetricsLister
	pinger service.Pinger
	logger *zap.Logger
}

// Возвращает новый экземпляр MetricHandler
func NewMetricHandler(
	writer service.MetricWriter,
	reader service.MetricReader,
	lister service.MetricsLister,
	pinger service.Pinger,
	logger *zap.Logger,
) *MetricHandler {
	return &MetricHandler{
		writer: writer,
		reader: reader,
		lister: lister,
		pinger: pinger,
		logger: logger,
	}
}
