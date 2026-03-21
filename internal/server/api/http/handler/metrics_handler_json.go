package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/models"
	"go.uber.org/zap"
)

// Сохраняет метрику из JSON body запроса
func (mh *MetricHandler) SaveMetricHandlerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		log := contextkey.LoggerFromContext(r.Context())
		log.Error("unexpected content type", zap.String("content-type", r.Header.Get("Content-Type")))
		mh.writeJSONError(w, "Expected JSON", http.StatusBadRequest)
		return
	}

	var req models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log := contextkey.LoggerFromContext(r.Context())
		log.Error("cannot decode request JSON body", zap.Error(err))
		mh.writeJSONError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	switch req.MType {
	case "gauge":
		mh.handleSaveGaugeJSON(w, r.Context(), req)
	case "counter":
		mh.handleSaveCounterJSON(w, r.Context(), req)
	default:
		log := contextkey.LoggerFromContext(r.Context())
		log.Error("unsupported metric type", zap.String("type", req.MType))
		mh.writeJSONError(w, "Unsupported metric type", http.StatusBadRequest)
	}
}

// Сохраняет метрики из JSON body запроса
func (mh *MetricHandler) SaveMetricsHandlerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		mh.logger.Error("unexpected content type", zap.String("content-type", r.Header.Get("Content-Type")))
		mh.writeJSONError(w, "expected JSON", http.StatusBadRequest)
		return
	}

	var req []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mh.logger.Error("cannot decode request JSON body", zap.Error(err))
		mh.writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	for _, metric := range req {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				mh.logger.Error("missing 'value' for gauge metric", zap.String("id", metric.ID))
				mh.writeJSONError(w, "missing value for gauge", http.StatusBadRequest)
				return
			}
			gauges[metric.ID] = *metric.Value
		case "counter":
			if metric.Delta == nil {
				mh.logger.Error("missing 'delta' for counter metric", zap.String("id", metric.ID))
				mh.writeJSONError(w, "missing value for counter", http.StatusBadRequest)
				return
			}
			counters[metric.ID] += *metric.Delta
		default:
			mh.logger.Error("unsupported metric type", zap.String("type", metric.MType))
			mh.writeJSONError(w, "unsupported metric type", http.StatusBadRequest)
		}
	}

	err := mh.writer.SaveAllMetrics(r.Context(), gauges, counters)
	if err != nil {
		mh.logger.Error("error processing metrics", zap.Any("gauges", gauges), zap.Any("counters", counters), zap.Error(err))
		mh.writeJSONError(w, "processing metrics failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) handleSaveGaugeJSON(w http.ResponseWriter, ctx context.Context, req models.Metrics) {
	if req.Value == nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("missing 'value' for gauge metric", zap.String("id", req.ID))
		mh.writeJSONError(w, "Missing value for gauge", http.StatusBadRequest)
		return
	}

	err := mh.writer.SaveGaugeMetric(ctx, req.ID, req.Value)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error processing gauge metric",
			zap.String("id", req.ID),
			zap.Float64("value", *req.Value),
			zap.Error(err))
		mh.writeJSONError(w, "Processing gauge metric failed", http.StatusBadRequest)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "gauge",
		Value: req.Value,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
}

func (mh *MetricHandler) handleSaveCounterJSON(w http.ResponseWriter, ctx context.Context, req models.Metrics) {
	if req.Delta == nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("missing 'delta' for counter metric", zap.String("id", req.ID))
		mh.writeJSONError(w, "Missing value for counter", http.StatusBadRequest)
		return
	}

	err := mh.writer.SaveCounterMetric(ctx, req.ID, req.Delta)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error processing counter metric",
			zap.String("id", req.ID),
			zap.Int64("delta", *req.Delta),
			zap.Error(err))
		mh.writeJSONError(w, "Processing counter metric failed", http.StatusBadRequest)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "counter",
		Delta: req.Delta,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
}

// Возвращает метрику по ID, указанному в JSON запросе
func (mh *MetricHandler) ValueMetricHandlerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		mh.logger.Error("unexpected content type", zap.String("content-type", r.Header.Get("Content-Type")))
		mh.writeJSONError(w, "expected JSON", http.StatusBadRequest)
		return
	}

	var req models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mh.logger.Error("cannot decode request JSON body", zap.Error(err))
		mh.writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	switch req.MType {
	case "gauge":
		mh.handleGetGaugeJSON(w, r.Context(), req)
	case "counter":
		mh.handleGetCounterJSON(w, r.Context(), req)
	default:
		mh.logger.Error("unsupported metric type",
			zap.Any("type", req.MType),
		)

		mh.writeJSONError(w, "unsupported metric type", http.StatusBadRequest)
	}
}

func (mh *MetricHandler) handleGetGaugeJSON(w http.ResponseWriter, ctx context.Context, req models.Metrics) {
	value, err := mh.reader.GetGaugeMetric(ctx, req.ID)
	if err != nil {
		mh.logger.Error("error getting gauge metric",
			zap.Any("name", req.ID),
			zap.Error(err),
		)
		mh.writeJSONError(w, "gauge not found", http.StatusNotFound)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "gauge",
		Value: &value,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
	mh.logger.Info("return gauge metric",
		zap.Any("name", req.ID),
		zap.Any("value", value),
	)

}

func (mh *MetricHandler) handleGetCounterJSON(w http.ResponseWriter, ctx context.Context, req models.Metrics) {
	value, err := mh.reader.GetCounterMetric(ctx, req.ID)
	if err != nil {
		mh.logger.Error("error getting counter metric",
			zap.Any("name", req.ID),
			zap.Error(err),
		)

		mh.writeJSONError(w, "counter not found", http.StatusNotFound)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "counter",
		Delta: &value,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
	mh.logger.Info("return counter metric",
		zap.Any("name", req.ID),
		zap.Any("value", value),
	)
}

func (mh *MetricHandler) writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (mh *MetricHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log := contextkey.LoggerFromContext(context.Background())
		log.Warn("failed to encode JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to encode JSON response"})
	}
}
