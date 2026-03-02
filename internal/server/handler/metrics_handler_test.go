package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/models"
	"go.uber.org/zap/zaptest"
)

func TestMetricHandler_SaveMetricHandlerJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		input          models.Metrics
		mockSave       func(ctx context.Context, name string, value *float64) error
		expectedStatus int
		expectedFields map[string]interface{}
	}{
		{
			name: "valid gauge",
			input: models.Metrics{
				ID:    "test_gauge",
				MType: "gauge",
				Value: floatPtr(12.34),
			},
			mockSave:       func(ctx context.Context, name string, value *float64) error { return nil },
			expectedStatus: http.StatusOK,
			expectedFields: map[string]interface{}{
				"id":    "test_gauge",
				"type":  "gauge",
				"value": 12.34,
			},
		},
		{
			name: "invalid gauge - missing value",
			input: models.Metrics{
				ID:    "test_gauge",
				MType: "gauge",
				Value: nil,
			},
			mockSave:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]interface{}{
				"error": "Missing value for gauge",
			},
		},
		{
			name: "save error",
			input: models.Metrics{
				ID:    "test_gauge",
				MType: "gauge",
				Value: floatPtr(12.34),
			},
			mockSave: func(ctx context.Context, name string, value *float64) error {
				return fmt.Errorf("db error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]interface{}{
				"error": "Processing gauge metric failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockMetricsService{
				saveGaugeFunc: tt.mockSave,
			}
			handler := NewMetricHandler(mockSvc, logger)

			reqBody, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			t.Logf("Sending JSON: %s", reqBody)

			req := httptest.NewRequest("POST", "/update/", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := contextkey.WithLogger(req.Context(), logger)
			req = req.WithContext(ctx)

			handler.SaveMetricHandlerJSON(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			t.Logf("response body: %s", body)

			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			for key, expected := range tt.expectedFields {
				actual, exists := result[key]
				if !exists {
					t.Errorf("response missing field: %s", key)
					continue
				}
				if key == "value" || key == "delta" {
					exp, ok1 := expected.(float64)
					act, ok2 := actual.(float64)
					if ok1 && ok2 {
						if act != exp {
							t.Errorf("field %s: expected %v, got %v", key, exp, act)
						}
					} else {
						t.Errorf("field %s: type mismatch, expected %T, got %T", key, expected, actual)
					}
				} else {
					if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
						t.Errorf("field %s: expected %v, got %v", key, expected, actual)
					}
				}
			}
		})
	}
}

func TestMetricHandler_ValueMetricHandlerJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		input          models.Metrics
		mockGet        func(ctx context.Context, name string) (float64, error)
		expectedStatus int
		expectedFields map[string]interface{}
	}{
		{
			name: "gauge found",
			input: models.Metrics{
				ID:    "test_gauge",
				MType: "gauge",
			},
			mockGet:        func(ctx context.Context, name string) (float64, error) { return 42.5, nil },
			expectedStatus: http.StatusOK,
			expectedFields: map[string]interface{}{
				"id":    "test_gauge",
				"type":  "gauge",
				"value": 42.5,
			},
		},
		{
			name: "gauge not found",
			input: models.Metrics{
				ID:    "missing",
				MType: "gauge",
			},
			mockGet:        func(ctx context.Context, name string) (float64, error) { return 0, fmt.Errorf("not found") },
			expectedStatus: http.StatusNotFound,
			expectedFields: map[string]interface{}{
				"error": "gauge not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockMetricsService{
				getGaugeFunc: tt.mockGet,
			}
			handler := NewMetricHandler(mockSvc, logger)

			reqBody, _ := json.Marshal(tt.input)
			t.Logf("request JSON: %s", reqBody)

			req := httptest.NewRequest("POST", "/value/", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ctx := contextkey.WithLogger(req.Context(), logger)
			req = req.WithContext(ctx)

			handler.ValueMetricHandlerJSON(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			t.Logf("response body: %s", body)

			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			for key, expected := range tt.expectedFields {
				actual, exists := result[key]
				if !exists {
					t.Errorf("missing field: %s", key)
					continue
				}
				if key == "value" {
					exp, ok1 := expected.(float64)
					act, ok2 := actual.(float64)
					if ok1 && ok2 {
						if act != exp {
							t.Errorf("value mismatch: expected %v, got %v", exp, act)
						}
					} else {
						t.Errorf("type mismatch for 'value': expected %T, got %T", expected, actual)
					}
				} else {
					if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
						t.Errorf("field %s: expected %v, got %v", key, expected, actual)
					}
				}
			}
		})
	}
}

func TestMetricHandler_SaveMetricHandlerPlain(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name            string
		metricType      string
		nameParam       string
		valueParam      string
		mockSaveGauge   func(ctx context.Context, name string, value *float64) error
		mockSaveCounter func(ctx context.Context, name string, delta *int64) error
		expectedStatus  int
	}{
		{
			name:           "valid gauge",
			metricType:     "gauge",
			nameParam:      "test_gauge",
			valueParam:     "3.14",
			mockSaveGauge:  func(ctx context.Context, name string, value *float64) error { return nil },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid gauge value",
			metricType:     "gauge",
			nameParam:      "test_gauge",
			valueParam:     "xyz",
			mockSaveGauge:  nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:            "valid counter",
			metricType:      "counter",
			nameParam:       "test_counter",
			valueParam:      "100",
			mockSaveCounter: func(ctx context.Context, name string, delta *int64) error { return nil },
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "invalid counter value",
			metricType:      "counter",
			nameParam:       "test_counter",
			valueParam:      "abc",
			mockSaveCounter: nil,
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:           "empty name",
			metricType:     "gauge",
			nameParam:      "",
			valueParam:     "1.0",
			mockSaveGauge:  nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid type",
			metricType:     "timer",
			nameParam:      "some_timer",
			valueParam:     "100",
			mockSaveGauge:  nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockSvc *mockMetricsService

			switch tt.metricType {
			case "gauge":
				mockSvc = &mockMetricsService{
					saveGaugeFunc: tt.mockSaveGauge,
				}
			case "counter":
				mockSvc = &mockMetricsService{
					saveCounterFunc: tt.mockSaveCounter,
				}
			default:
				mockSvc = &mockMetricsService{}
			}

			handler := NewMetricHandler(mockSvc, logger)

			r := httptest.NewRequest("POST", "/", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.metricType)
			rctx.URLParams.Add("name", tt.nameParam)
			rctx.URLParams.Add("value", tt.valueParam)

			ctx := contextkey.WithLogger(r.Context(), logger)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
			r = r.WithContext(ctx)

			handler.SaveMetricHandler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricHandler_GetMetricHandlerPlain(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		metricType     string
		nameParam      string
		mockGetGauge   func(ctx context.Context, name string) (float64, error)
		mockGetCounter func(ctx context.Context, name string) (int64, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "gauge found",
			metricType:     "gauge",
			nameParam:      "cpu_usage",
			mockGetGauge:   func(ctx context.Context, name string) (float64, error) { return 1.99, nil },
			expectedStatus: http.StatusOK,
			expectedBody:   "1.99",
		},
		{
			name:           "gauge not found",
			metricType:     "gauge",
			nameParam:      "missing",
			mockGetGauge:   func(ctx context.Context, name string) (float64, error) { return 0, fmt.Errorf("not found") },
			expectedStatus: http.StatusNotFound,
			expectedBody:   "gauge not found",
		},
		{
			name:           "counter found",
			metricType:     "counter",
			nameParam:      "requests",
			mockGetCounter: func(ctx context.Context, name string) (int64, error) { return 42, nil },
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name:           "counter not found",
			metricType:     "counter",
			nameParam:      "missing",
			mockGetCounter: func(ctx context.Context, name string) (int64, error) { return 0, fmt.Errorf("not found") },
			expectedStatus: http.StatusNotFound,
			expectedBody:   "counter not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockSvc *mockMetricsService

			switch tt.metricType {
			case "gauge":
				mockSvc = &mockMetricsService{
					getGaugeFunc: tt.mockGetGauge,
				}
			case "counter":
				mockSvc = &mockMetricsService{
					getCounterFunc: tt.mockGetCounter,
				}
			default:
				mockSvc = &mockMetricsService{}
			}

			handler := NewMetricHandler(mockSvc, logger)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.metricType)
			rctx.URLParams.Add("name", tt.nameParam)

			ctx := contextkey.WithLogger(r.Context(), logger)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
			r = r.WithContext(ctx)

			handler.GetMetricHandler(w, r)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := strings.TrimSpace(string(body))

			if tt.expectedBody != "" && !strings.Contains(bodyStr, tt.expectedBody) {
				t.Errorf("expected body containing %q, got %q", tt.expectedBody, bodyStr)
			}
		})
	}
}

func TestMetricHandler_CollectMetricsHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockSvc := &mockMetricsService{
		listAllFunc: func(ctx context.Context) (map[string]float64, map[string]int64) {
			return map[string]float64{"temp": 36.6}, map[string]int64{"calls": 1000}
		},
	}

	handler := NewMetricHandler(mockSvc, logger)
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	ctx := contextkey.WithLogger(req.Context(), logger)
	req = req.WithContext(ctx)

	handler.CollectMetricsHandler(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, html.EscapeString("temp: 36.6")) ||
		!strings.Contains(bodyStr, html.EscapeString("calls: 1000")) {
		t.Errorf("expected metrics in HTML, got: %s", bodyStr)
	}
}

func TestMetricHandler_DBHealthCheckHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		pingResult     bool
		expectedStatus int
	}{
		{name: "db healthy", pingResult: true, expectedStatus: http.StatusOK},
		{name: "db unreachable", pingResult: false, expectedStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockMetricsService{
				pingFunc: func(ctx context.Context) bool { return tt.pingResult },
			}
			handler := NewMetricHandler(mockSvc, logger)

			req := httptest.NewRequest("GET", "/ping", nil)
			w := httptest.NewRecorder()

			ctx := contextkey.WithLogger(req.Context(), logger)
			req = req.WithContext(ctx)

			handler.DBHealthCheckHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func floatPtr(f float64) *float64 { return &f }
func int64Ptr(i int64) *int64     { return &i }
