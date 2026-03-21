package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
)

// Сохраняет метрику из plain body запроса
func (mh *MetricHandler) SaveMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	log := contextkey.LoggerFromContext(r.Context())

	if !isValidMetricType(mType) {
		log.Error("invalid metric type", zap.String("type", mType))
		mh.writeTextError(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		log.Error("invalid metric name", zap.String("name", mName))
		mh.writeTextError(w, "invalid metric name", http.StatusBadRequest)
		return
	}

	switch mType {
	case "gauge":
		mh.handleSaveGaugePlain(w, r.Context(), mName, mValue)
	case "counter":
		mh.handleSaveCounterPlain(w, r.Context(), mName, mValue)
	}
}

func (mh *MetricHandler) handleSaveGaugePlain(w http.ResponseWriter, ctx context.Context, name, value string) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error parsing gauge value", zap.String("name", name), zap.String("value", value), zap.Error(err))
		mh.writeTextError(w, "invalid gauge value", http.StatusBadRequest)
		return
	}

	err = mh.writer.SaveGaugeMetric(ctx, name, &floatValue)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error saving gauge metric", zap.String("name", name), zap.Float64("value", floatValue), zap.Error(err))
		mh.writeTextError(w, "failed to save gauge", http.StatusBadRequest)
		return
	}

	mh.writeTextResponse(w, "", http.StatusOK)
}

func (mh *MetricHandler) handleSaveCounterPlain(w http.ResponseWriter, ctx context.Context, name, value string) {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error parsing counter value", zap.String("name", name), zap.String("value", value), zap.Error(err))
		mh.writeTextError(w, "invalid counter value", http.StatusBadRequest)
		return
	}

	err = mh.writer.SaveCounterMetric(ctx, name, &intValue)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error saving counter metric", zap.String("name", name), zap.Int64("value", intValue), zap.Error(err))
		mh.writeTextError(w, "failed to save counter", http.StatusBadRequest)
		return
	}

	mh.writeTextResponse(w, "", http.StatusOK)
}

// Возвращает метрику по ID, указанному в URL
func (mh *MetricHandler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	log := contextkey.LoggerFromContext(r.Context())

	if !isValidMetricType(mType) {
		log.Error("invalid metric type", zap.String("type", mType))
		mh.writeTextError(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		log.Error("invalid metric name", zap.String("name", mName))
		mh.writeTextError(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	switch mType {
	case "gauge":
		mh.handleGetGaugePlain(w, r.Context(), mName)
	case "counter":
		mh.handleGetCounterPlain(w, r.Context(), mName)
	}
}

func (mh *MetricHandler) handleGetGaugePlain(w http.ResponseWriter, ctx context.Context, name string) {
	value, err := mh.reader.GetGaugeMetric(ctx, name)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error getting gauge metric", zap.String("name", name), zap.Error(err))
		mh.writeTextError(w, "gauge not found", http.StatusNotFound)
		return
	}

	mh.writeTextResponse(w, fmt.Sprintf("%g", value), http.StatusOK)
}

func (mh *MetricHandler) handleGetCounterPlain(w http.ResponseWriter, ctx context.Context, name string) {
	value, err := mh.reader.GetCounterMetric(ctx, name)
	if err != nil {
		log := contextkey.LoggerFromContext(ctx)
		log.Error("error getting counter metric", zap.String("name", name), zap.Error(err))
		mh.writeTextError(w, "counter not found", http.StatusNotFound)
		return
	}

	mh.writeTextResponse(w, fmt.Sprintf("%d", value), http.StatusOK)
}

// Возвращает все метрики в HTML формате
func (mh *MetricHandler) CollectMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauges, counters := mh.lister.ListAllMetrics(r.Context())
	renderMetricsHTML(w, gauges, counters)
}

// Проверяет достуность БД
func (mh *MetricHandler) DBHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log := contextkey.LoggerFromContext(r.Context())
	if mh.pinger == nil {
		log.Error("service not available")
		http.Error(w, "service not available", http.StatusInternalServerError)
		return
	}
	if !mh.pinger.Ping(r.Context()) {
		log.Error("database unreachable")
		http.Error(w, "database unreachable", http.StatusInternalServerError)
		return
	}
	log.Info("database health check OK")
	w.WriteHeader(http.StatusOK)
}

func isValidMetricType(metricType string) bool {
	return metricType == "gauge" || metricType == "counter"
}

func isValidMetricName(mName string) bool {
	return len(mName) > 0
}

func (mh *MetricHandler) writeTextError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(w, message, status)
}

func (mh *MetricHandler) writeTextResponse(w http.ResponseWriter, body string, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
}
