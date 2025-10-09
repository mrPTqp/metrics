package handler

import (
	"net/http"
	"strconv"
	"strings"

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

func (mh *MetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathParts) != 4 {
		http.Error(w, "Invalid URL format. Expected: /update/<type>/<name>/<value>", http.StatusNotFound)
		return
	}

	mType := pathParts[1]
	mName := pathParts[2]
	mValue := pathParts[3]

	if !isValidMetricType(mType) {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !isValidMetricName(mName) {
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	switch mType {
	case "gauge":
		floatValue, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
		}
		err = mh.service.ProcessGaugeMetric(mName, floatValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "counter":
		intValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		err = mh.service.ProcessCounterMetric(mName, intValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func isValidMetricType(metricType string) bool {
	metricType = strings.ToLower(metricType)
	if metricType == "gauge" || metricType == "counter" {
		return true
	}
	return false
}

func isValidMetricName(mName string) bool {
	return len(mName) > 0
}
