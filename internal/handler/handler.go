package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mrPTqp/metrics/internal/models"
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

func (mh *MetricHandler) SaveMetricHandlerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		mh.logger.Error("unexpected content type", zap.String("content-type", r.Header.Get("Content-Type")))
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "expected JSON"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		mh.logger.Error("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "invalid JSON body"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	switch req.MType {
	case "gauge":
		if req.Value == nil {
			mh.logger.Error("missing 'value' for gauge metric %s", req.ID)
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"error": "missing value for gauge"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		err := mh.service.SaveGaugeMetric(strings.ToLower(req.ID), req.Value)
		if err != nil {
			mh.logger.Errorf("Error processing gauge metric: %s - %v", req.ID, req.Value, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"error": "processing gauge metric failed"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp := models.Metrics{
			ID:    req.ID,
			MType: "gauge",
			Value: req.Value,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			mh.logger.Warn("Failed to encode JSON response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": "failed to encode JSON response"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	case "counter":
		if req.Delta == nil {
			mh.logger.Error("missing 'delta' for counter metric %s", req.ID)
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"error": "missing value for counter"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		err := mh.service.SaveCounterMetric(strings.ToLower(req.ID), req.Delta)
		if err != nil {
			mh.logger.Error("Error processing counter metric: %s - %v", req.ID, req.Delta, zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"error": "processing counter metric failed"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp := models.Metrics{
			ID:    req.ID,
			MType: "counter",
			Delta: req.Delta,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			mh.logger.Warn("Failed to encode JSON response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": "failed to encode JSON response"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	default:
		mh.logger.Error("Unsupported metric type: %s", req.MType)
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "unsupported metric type"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
}

func (mh *MetricHandler) ValueMetricHandlerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		mh.logger.Error("unexpected content type", zap.String("content-type", r.Header.Get("Content-Type")))
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "expected JSON"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		mh.logger.Error("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "invalid JSON body"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	switch req.MType {
	case "gauge":
		mValue, err := mh.service.GetGaugeMetric(strings.ToLower(req.ID))
		if err != nil {
			mh.logger.Error("Error getting gauge metric: %s", req.ID, zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			resp := map[string]string{"error": "gauge not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp := models.Metrics{
			ID:    req.ID,
			MType: "gauge",
			Value: &mValue,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			mh.logger.Error("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": "error encoding response"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		mh.logger.Info("return gauge metric %s value %f", req.ID, mValue)

	case "counter":
		mValue, err := mh.service.GetCounterMetric(strings.ToLower(req.ID))
		if err != nil {
			mh.logger.Error("Error getting counter metric: %s", req.ID, zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			resp := map[string]string{"error": "counter not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp := models.Metrics{
			ID:    req.ID,
			MType: "counter",
			Delta: &mValue,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			mh.logger.Error("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": "error encoding response"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		mh.logger.Info("return counter metric %s value %d", req.ID, mValue)

	default:
		mh.logger.Error("Unsupported metric type: %s", req.MType)
		http.Error(w, "Unsupported metric type", http.StatusBadRequest)
		return
	}
}

func (mh *MetricHandler) CollectMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauges, counters := mh.service.ListAllMetrics()
	renderMetricsHTML(w, gauges, counters)
}

func (mh *MetricHandler) SaveMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := strings.ToLower(chi.URLParam(r, "type"))
	mName := strings.ToLower(chi.URLParam(r, "name"))
	mValue := chi.URLParam(r, "value")

	if !isValidMetricType(mType) {
		mh.logger.Error("Invalid metric type: %s", mType)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		mh.logger.Error("Invalid metric name: %s", mName)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch mType {
	case "gauge":
		floatValue, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			mh.logger.Error("Error parsing gauge value: %s - %v", mName, mValue, zap.Error(err))
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		err = mh.service.SaveGaugeMetric(mName, &floatValue)
		if err != nil {
			mh.logger.Error("Error processing gauge metric: %s - %v", mName, mValue, zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "counter":
		intValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			mh.logger.Error("Error parsing counter value: %s - %v", mName, mValue, zap.Error(err))
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		err = mh.service.SaveCounterMetric(mName, &intValue)
		if err != nil {
			mh.logger.Error("Error processing counter metric: %s - %v", mName, mValue, zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := strings.ToLower(chi.URLParam(r, "type"))
	mName := strings.ToLower(chi.URLParam(r, "name"))

	if !isValidMetricType(mType) {
		mh.logger.Error("Invalid metric type: %s", mType)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		mh.logger.Error("Invalid metric name: %s", mName)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch mType {
	case "gauge":
		mValue, err := mh.service.GetGaugeMetric(mName)
		if err != nil {
			mh.logger.Error("Error getting gauge metric: %s", mName, zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%g", mValue)))
	case "counter":
		mValue, err := mh.service.GetCounterMetric(mName)
		if err != nil {
			mh.logger.Error("Error getting counter metric: %s", mName, zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%d", mValue)))
	default:
		mh.logger.Error("Unsupported metric type: %s", mType)
		http.Error(w, "Unsupported metric type", http.StatusBadRequest)
		return
	}
}

func isValidMetricType(metricType string) bool {
	metricType = strings.ToLower(metricType)
	return metricType == "gauge" || metricType == "counter"
}

func isValidMetricName(mName string) bool {
	return len(mName) > 0
}
