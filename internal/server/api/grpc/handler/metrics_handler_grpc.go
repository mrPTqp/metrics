package handler

import (
	"context"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/proto"
	"github.com/mrPTqp/metrics/internal/server/service"
)

// MetricsHandler обрабатывает gRPC запросы
type MetricsHandler struct {
	proto.UnimplementedMetricsServer
	writer service.MetricWriter
	logger *zap.Logger
}

// NewMetricsHandler создает новый MetricsHandler
func NewMetricsHandler(writer service.MetricWriter, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		writer: writer,
		logger: logger,
	}
}

// UpdateMetrics обновляет метрики
func (h *MetricsHandler) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for _, metric := range req.Metrics {
		switch metric.Type {
		case proto.Metric_GAUGE:
			if metric.Value != nil {
				gauges[metric.Id] = *metric.Value
			}
		case proto.Metric_COUNTER:
			if metric.Delta != nil {
				counters[metric.Id] = *metric.Delta
			}
		}
	}

	if err := h.writer.SaveAllMetrics(ctx, gauges, counters); err != nil {
		h.logger.Error("failed to save metrics", zap.Error(err))
		return nil, err
	}

	h.logger.Info("metrics updated via gRPC")
	return &proto.UpdateMetricsResponse{}, nil
}

