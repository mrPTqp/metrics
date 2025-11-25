package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (mh *MetricHandler) SaveMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	if !isValidMetricType(mType) {
		mh.logger.Error("Invalid metric type: %s", mType)
		mh.writeTextError(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		mh.logger.Error("Invalid metric name: %s", mName)
		mh.writeTextError(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	switch mType {
	case "gauge":
		mh.handleSaveGaugePlain(w, mName, mValue)
	case "counter":
		mh.handleSaveCounterPlain(w, mName, mValue)
	}
}

func (mh *MetricHandler) handleSaveGaugePlain(w http.ResponseWriter, name, value string) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		mh.logger.Error("Error parsing gauge value: %s - %s", name, value, zap.Error(err))
		mh.writeTextError(w, "Invalid gauge value", http.StatusBadRequest)
		return
	}

	err = mh.service.SaveGaugeMetric(name, &floatValue)
	if err != nil {
		mh.logger.Error("Error processing gauge metric: %s - %v", name, floatValue, zap.Error(err))
		mh.writeTextError(w, "Failed to save gauge", http.StatusBadRequest)
		return
	}

	mh.writeTextResponse(w, "", http.StatusOK)
}

func (mh *MetricHandler) handleSaveCounterPlain(w http.ResponseWriter, name, value string) {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		mh.logger.Error("Error parsing counter value: %s - %s", name, value, zap.Error(err))
		mh.writeTextError(w, "Invalid counter value", http.StatusBadRequest)
		return
	}

	err = mh.service.SaveCounterMetric(name, &intValue)
	if err != nil {
		mh.logger.Error("Error processing counter metric: %s - %d", name, intValue, zap.Error(err))
		mh.writeTextError(w, "Failed to save counter", http.StatusBadRequest)
		return
	}

	mh.writeTextResponse(w, "", http.StatusOK)
}

func (mh *MetricHandler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	if !isValidMetricType(mType) {
		mh.logger.Error("Invalid metric type: %s", mType)
		mh.writeTextError(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		mh.logger.Error("Invalid metric name: %s", mName)
		mh.writeTextError(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	switch mType {
	case "gauge":
		mh.handleGetGaugePlain(w, mName)
	case "counter":
		mh.handleGetCounterPlain(w, mName)
	}
}

func (mh *MetricHandler) handleGetGaugePlain(w http.ResponseWriter, name string) {
	value, err := mh.service.GetGaugeMetric(name)
	if err != nil {
		mh.logger.Error("Error getting gauge metric: %s", name, zap.Error(err))
		mh.writeTextError(w, "gauge not found", http.StatusNotFound)
		return
	}

	mh.writeTextResponse(w, fmt.Sprintf("%g", value), http.StatusOK)
}

func (mh *MetricHandler) handleGetCounterPlain(w http.ResponseWriter, name string) {
	value, err := mh.service.GetCounterMetric(name)
	if err != nil {
		mh.logger.Error("Error getting counter metric: %s", name, zap.Error(err))
		mh.writeTextError(w, "counter not found", http.StatusNotFound)
		return
	}

	mh.writeTextResponse(w, fmt.Sprintf("%d", value), http.StatusOK)
}

func (mh *MetricHandler) CollectMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauges, counters := mh.service.ListAllMetrics()
	renderMetricsHTML(w, gauges, counters)
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

func (mh *MetricHandler) DBHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    if mh.service == nil {
        http.Error(w, "service not available", http.StatusInternalServerError)
        return
    }
    if !mh.service.Ping() {
        http.Error(w, "database unreachable", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}