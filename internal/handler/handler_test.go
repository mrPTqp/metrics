package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetricsService struct {
	processGaugeCalls   []string
	processCounterCalls []string
	getGaugeValue       float64
	getCounterValue     int64
}

func (m *mockMetricsService) ProcessGaugeMetric(name string, value float64) error {
	m.processGaugeCalls = append(m.processGaugeCalls, name)
	return nil
}

func (m *mockMetricsService) ProcessCounterMetric(name string, value int64) error {
	m.processCounterCalls = append(m.processCounterCalls, name)
	return nil
}

func (m *mockMetricsService) GetGaugeMetric(name string) (float64, error) {
	return m.getGaugeValue, nil
}

func (m *mockMetricsService) GetCounterMetric(name string) (int64, error) {
	return m.getCounterValue, nil
}

func setupRouter(handler http.HandlerFunc) *chi.Mux {
	r := chi.NewRouter()
	// Make MethodNotAllowed return body consistent with handler's own check
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.Post("/metrics/{type}/{name}/{value}", handler)
	r.Get("/metrics/{type}/{name}", handler)
	return r
}

func TestSaveMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		metricType     string
		metricName     string
		metricValue    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "InvalidMethod",
			method:         http.MethodGet,
			metricType:     "gauge",
			metricName:     "test",
			metricValue:    "123.45",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "InvalidMetricType",
			method:         http.MethodPost,
			metricType:     "invalid",
			metricName:     "test",
			metricValue:    "123.45",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
		{
			name:           "InvalidMetricName",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "",
			metricValue:    "123.45",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric name\n",
		},
		{
			name:           "InvalidGaugeValue",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "test",
			metricValue:    "invalid",
			expectedStatus: http.StatusBadRequest,
			// handler logs the parse error but returns this body
			expectedBody: "Invalid gauge value\n",
		},
		{
			name:           "InvalidCounterValue",
			method:         http.MethodPost,
			metricType:     "counter",
			metricName:     "test",
			metricValue:    "notanumber",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid counter value\n",
		},
		{
			name:           "ValidGauge",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "cpu",
			metricValue:    "123.45",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "ValidCounter",
			method:         http.MethodPost,
			metricType:     "counter",
			metricName:     "requests",
			metricValue:    "456",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockMetricsService{}
			handler := NewMetricHandler(mockService)

			r := setupRouter(handler.SaveMetricHandler)

			req, err := http.NewRequest(tt.method, fmt.Sprintf("/metrics/%s/%s/%s",
				tt.metricType, tt.metricName, tt.metricValue), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			if tt.expectedStatus == http.StatusOK {
				switch tt.metricType {
				case "gauge":
					assert.Contains(t, mockService.processGaugeCalls, tt.metricName)
				case "counter":
					assert.Contains(t, mockService.processCounterCalls, tt.metricName)
				}
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		metricType     string
		metricName     string
		setupMock      func(*mockMetricsService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "InvalidMethod",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "test",
			setupMock:      func(m *mockMetricsService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "InvalidMetricType",
			method:         http.MethodGet,
			metricType:     "invalid",
			metricName:     "test",
			setupMock:      func(m *mockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
		{
			name:           "InvalidMetricName",
			method:         http.MethodGet,
			metricType:     "gauge",
			metricName:     "",
			setupMock:      func(m *mockMetricsService) {},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:       "ValidGauge",
			method:     http.MethodGet,
			metricType: "gauge",
			metricName: "temp",
			setupMock: func(m *mockMetricsService) {
				m.getGaugeValue = 123.45
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "123.450000",
		},
		{
			name:       "ValidCounter",
			method:     http.MethodGet,
			metricType: "counter",
			metricName: "hits",
			setupMock: func(m *mockMetricsService) {
				m.getCounterValue = 456
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockMetricsService{}
			tt.setupMock(mockService)

			handler := NewMetricHandler(mockService)

			r := setupRouter(handler.GetMetricHandler)

			req, err := http.NewRequest(tt.method, fmt.Sprintf("/metrics/%s/%s",
				tt.metricType, tt.metricName), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
