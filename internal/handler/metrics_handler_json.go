package handler

import (
	"encoding/json"
	"github.com/mrPTqp/metrics/internal/models"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func (mh *MetricHandler) SaveMetricHandlerJSON(w http.ResponseWriter, r *http.Request) {
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
		mh.handleSaveGaugeJSON(w, req)
	case "counter":
		mh.handleSaveCounterJSON(w, req)
	default:
		mh.logger.Error("Unsupported metric type: %s", req.MType)
		mh.writeJSONError(w, "unsupported metric type", http.StatusBadRequest)
	}
}

func (mh *MetricHandler) handleSaveGaugeJSON(w http.ResponseWriter, req models.Metrics) {
	if req.Value == nil {
		mh.logger.Error("missing 'value' for gauge metric %s", req.ID)
		mh.writeJSONError(w, "missing value for gauge", http.StatusBadRequest)
		return
	}

	err := mh.service.SaveGaugeMetric(strings.ToLower(req.ID), req.Value)
	if err != nil {
		mh.logger.Error("Error processing gauge metric: %s - %v", req.ID, *req.Value, zap.Error(err))
		mh.writeJSONError(w, "processing gauge metric failed", http.StatusBadRequest)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "gauge",
		Value: req.Value,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
}

func (mh *MetricHandler) handleSaveCounterJSON(w http.ResponseWriter, req models.Metrics) {
	if req.Delta == nil {
		mh.logger.Error("missing 'delta' for counter metric %s", req.ID)
		mh.writeJSONError(w, "missing value for counter", http.StatusBadRequest)
		return
	}

	err := mh.service.SaveCounterMetric(strings.ToLower(req.ID), req.Delta)
	if err != nil {
		mh.logger.Error("Error processing counter metric: %s - %v", req.ID, *req.Delta, zap.Error(err))
		mh.writeJSONError(w, "processing counter metric failed", http.StatusBadRequest)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "counter",
		Delta: req.Delta,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
}

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
		mh.handleGetGaugeJSON(w, req)
	case "counter":
		mh.handleGetCounterJSON(w, req)
	default:
		mh.logger.Error("Unsupported metric type: %s", req.MType)
		mh.writeJSONError(w, "unsupported metric type", http.StatusBadRequest)
	}
}

func (mh *MetricHandler) handleGetGaugeJSON(w http.ResponseWriter, req models.Metrics) {
	value, err := mh.service.GetGaugeMetric(strings.ToLower(req.ID))
	if err != nil {
		mh.logger.Error("Error getting gauge metric: %s", req.ID, zap.Error(err))
		mh.writeJSONError(w, "gauge not found", http.StatusNotFound)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "gauge",
		Value: &value,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
	mh.logger.Info("return gauge metric %s value %f", req.ID, value)
}

func (mh *MetricHandler) handleGetCounterJSON(w http.ResponseWriter, req models.Metrics) {
	value, err := mh.service.GetCounterMetric(strings.ToLower(req.ID))
	if err != nil {
		mh.logger.Error("Error getting counter metric: %s", req.ID, zap.Error(err))
		mh.writeJSONError(w, "counter not found", http.StatusNotFound)
		return
	}

	resp := models.Metrics{
		ID:    req.ID,
		MType: "counter",
		Delta: &value,
	}
	mh.writeJSONResponse(w, resp, http.StatusOK)
	mh.logger.Info("return counter metric %s value %d", req.ID, value)
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
		mh.logger.Warn("Failed to encode JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to encode JSON response"})
	}
}
