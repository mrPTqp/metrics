package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mrPTqp/metrics/internal/service"
)

type MetricHandler struct {
	service service.MetricsService
}

func NewMetricHandler(service service.MetricsService) *MetricHandler {
	return &MetricHandler{
		service: service,
	}
}

func (mh *MetricHandler) SaveMetricHandler(w http.ResponseWriter, r *http.Request) {
	mType := strings.ToLower(chi.URLParam(r, "type"))
	mName := strings.ToLower(chi.URLParam(r, "name"))
	mValue := chi.URLParam(r, "value")

	if !isValidMetricType(mType) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch mType {
	case "gauge":
		floatValue, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			log.Printf("Error parsing gauge value: %v", err)
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		err = mh.service.SaveGaugeMetric(mName, floatValue)
		if err != nil {
			log.Printf("Error processing gauge metric: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "counter":
		intValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			log.Printf("Error parsing counter value: %v", err)
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		err = mh.service.SaveCounterMetric(mName, intValue)
		if err != nil {
			log.Printf("Error processing counter metric: %v", err)
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
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch mType {
	case "gauge":
		mValue, err := mh.service.GetGaugeMetric(mName)
		if err != nil {
			log.Printf("Error getting gauge metric: %v", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%g", mValue)))
	case "counter":
		mValue, err := mh.service.GetCounterMetric(mName)
		if err != nil {
			log.Printf("Error getting counter metric: %v", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%d", mValue)))
	default:
		log.Printf("Unsupported metric type: %s", mType)
		http.Error(w, "Unsupported metric type", http.StatusBadRequest)
		return
	}
}

func (mh *MetricHandler) CollectMetricsHandler(w http.ResponseWriter, r *http.Request) {
	gauges, counters := mh.service.ListAllMetrics()

	renderMetricsHTML(w, gauges, counters)
}

func isValidMetricType(metricType string) bool {
	metricType = strings.ToLower(metricType)
	return metricType == "gauge" || metricType == "counter"
}

func isValidMetricName(mName string) bool {
	return len(mName) > 0
}
