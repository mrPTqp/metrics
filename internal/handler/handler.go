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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	mType := strings.ToLower(chi.URLParam(r, "type"))
	mName := strings.ToLower(chi.URLParam(r, "name"))
	mValue := chi.URLParam(r, "value")

	if !isValidMetricType(mType) {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	if !isValidMetricName(mName) {
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	switch mType {
	case "gauge":
		floatValue, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			log.Printf("Error parsing gauge value: %v", err)
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			return
		}
		err = mh.service.ProcessGaugeMetric(mName, floatValue)
		if err != nil {
			log.Printf("Error processing gauge metric: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			return
		}
	case "counter":
		intValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			log.Printf("Error parsing counter value: %v", err)
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			return
		}
		err = mh.service.ProcessCounterMetric(mName, intValue)
		if err != nil {
			log.Printf("Error processing counter metric: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	mType := strings.ToLower(chi.URLParam(r, "type"))
	mName := strings.ToLower(chi.URLParam(r, "name"))

	if !isValidMetricType(mType) {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	if !isValidMetricName(mName) {
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	switch mType {
	case "gauge":
		mValue, err := mh.service.GetGaugeMetric(mName)
		if err != nil {
			log.Printf("Error getting gauge metric: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			return
		}
		w.Write([]byte(fmt.Sprintf("%f", mValue)))
	case "counter":
		mValue, err := mh.service.GetCounterMetric(mName)
		if err != nil {
			log.Printf("Error getting counter metric: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			return
		}
		w.Write([]byte(fmt.Sprintf("%d", mValue)))
	default:
		log.Printf("Unsupported metric type: %s", mType)
		http.Error(w, "Unsupported metric type", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func isValidMetricType(metricType string) bool {
	metricType = strings.ToLower(metricType)
	return metricType == "gauge" || metricType == "counter"
}

func isValidMetricName(mName string) bool {
	return len(mName) > 0
}
