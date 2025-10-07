package handler

import (
	"github.com/mrPTqp/metrics/internal/service"
	"net/http"
	"strings"
)

type MetricHandler struct {
	ms *service.MetricsService
}

func NewMetricHandler(ms *service.MetricsService) *MetricHandler {
	return &MetricHandler{
		ms: ms,
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

	metricType := pathParts[1]
	metricName := pathParts[2]
	metricValue := pathParts[3]

	if !isValidMetricType(metricType) {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	err := mh.ms.ProcessMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
