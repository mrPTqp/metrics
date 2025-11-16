package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockMetricsService struct {
	saveGaugeFunc      func(name string, value *float64) error
	saveCounterFunc    func(name string, value *int64) error
	getGaugeFunc       func(name string) (float64, error)
	getCounterFunc     func(name string) (int64, error)
	listAllFunc        func() (map[string]float64, map[string]int64)
	saveAllMetricsFunc func(gauges map[string]float64, counters map[string]int64) error
	pingFunc           func() bool
}

func (m *mockMetricsService) SaveGaugeMetric(name string, value *float64) error {
	if m.saveGaugeFunc != nil {
		return m.saveGaugeFunc(name, value)
	}
	return nil
}

func (m *mockMetricsService) SaveCounterMetric(name string, value *int64) error {
	if m.saveCounterFunc != nil {
		return m.saveCounterFunc(name, value)
	}
	return nil
}

func (m *mockMetricsService) GetGaugeMetric(name string) (float64, error) {
	if m.getGaugeFunc != nil {
		return m.getGaugeFunc(name)
	}
	return 0, nil
}

func (m *mockMetricsService) GetCounterMetric(name string) (int64, error) {
	if m.getCounterFunc != nil {
		return m.getCounterFunc(name)
	}
	return 0, nil
}

func (m *mockMetricsService) ListAllMetrics() (map[string]float64, map[string]int64) {
	if m.listAllFunc != nil {
		return m.listAllFunc()
	}
	return map[string]float64{}, map[string]int64{}
}

func (m *mockMetricsService) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	if m.saveAllMetricsFunc != nil {
		return m.saveAllMetricsFunc(gauges, counters)
	}
	return nil
}

func (m *mockMetricsService) Ping() bool {
	if m.pingFunc != nil {
		return m.pingFunc()
	}
	return true
}

func ptr[T any](v T) *T { return &v }

func TestSaveMetricHandlerJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name           string
		reqBody        *models.Metrics
		contentType    string
		mockService    *mockMetricsService
		expectedStatus int
		expectSaved    func(t *testing.T, name string, value *float64, delta *int64)
	}{
		{
			name: "SaveGauge_Success",
			reqBody: &models.Metrics{
				ID:    "cpu",
				MType: "gauge",
				Value: ptr(3.14159),
			},
			contentType: "application/json",
			mockService: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					assert.Equal(t, "cpu", name)
					assert.NotNil(t, value)
					assert.InDelta(t, 3.14159, *value, 0.0001)
					return nil
				},
			},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T, name string, value *float64, delta *int64) {
				assert.Equal(t, "cpu", name)
				assert.NotNil(t, value)
				assert.InDelta(t, 3.14159, *value, 0.0001)
			},
		},
		{
			name: "SaveCounter_Success",
			reqBody: &models.Metrics{
				ID:    "hits",
				MType: "counter",
				Delta: ptr(int64(42)),
			},
			contentType: "application/json",
			mockService: &mockMetricsService{
				saveCounterFunc: func(name string, value *int64) error {
					assert.Equal(t, "hits", name)
					assert.NotNil(t, value)
					assert.Equal(t, int64(42), *value)
					return nil
				},
			},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T, name string, value *float64, delta *int64) {
				assert.Equal(t, "hits", name)
				assert.NotNil(t, delta)
				assert.Equal(t, int64(42), *delta)
			},
		},
		{
			name: "InvalidContentType",
			reqBody: &models.Metrics{
				ID:    "cpu",
				MType: "gauge",
				Value: ptr(1.23),
			},
			contentType:    "text/plain",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "InvalidJSON",
			reqBody:        nil,
			contentType:    "application/json",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "InvalidMetricType",
			reqBody: &models.Metrics{
				ID:    "test",
				MType: "unknown",
			},
			contentType:    "application/json",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "MissingGaugeValue",
			reqBody: &models.Metrics{
				ID:    "cpu",
				MType: "gauge",
			},
			contentType:    "application/json",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "MissingCounterDelta",
			reqBody: &models.Metrics{
				ID:    "hits",
				MType: "counter",
			},
			contentType:    "application/json",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceError",
			reqBody: &models.Metrics{
				ID:    "cpu",
				MType: "gauge",
				Value: ptr(1.23),
			},
			contentType: "application/json",
			mockService: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					return errors.New("service error")
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "NameLowercaseConversion",
			reqBody: &models.Metrics{
				ID:    "CPU",
				MType: "gauge",
				Value: ptr(1.23),
			},
			contentType: "application/json",
			mockService: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					assert.Equal(t, "cpu", name)
					return nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "EmptyBody",
			reqBody:        nil,
			contentType:    "application/json",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mockService, logger)
			r := chi.NewRouter()
			r.Post("/update", mh.SaveMetricHandlerJSON)

			var body []byte
			if tt.reqBody != nil {
				body, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectSaved != nil && resp.StatusCode == http.StatusOK {
				respBody, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				var result models.Metrics
				assert.NoError(t, json.Unmarshal(respBody, &result))
				tt.expectSaved(t, result.ID, result.Value, result.Delta)
			}
		})
	}
}

func TestValueMetricHandlerJSON(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name           string
		reqBody        models.Metrics
		mockService    *mockMetricsService
		expectedStatus int
		expectedValue  *float64
		expectedDelta  *int64
	}{
		{
			name: "GetGauge_Success",
			reqBody: models.Metrics{
				ID:    "temp",
				MType: "gauge",
			},
			mockService: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					assert.Equal(t, "temp", name)
					return 25.75, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedValue:  ptr(25.75),
		},
		{
			name: "GetCounter_Success",
			reqBody: models.Metrics{
				ID:    "hits",
				MType: "counter",
			},
			mockService: &mockMetricsService{
				getCounterFunc: func(name string) (int64, error) {
					assert.Equal(t, "hits", name)
					return 42, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedDelta:  ptr(int64(42)),
		},
		{
			name: "GaugeNotFound",
			reqBody: models.Metrics{
				ID:    "missing",
				MType: "gauge",
			},
			mockService: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					return 0, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "CounterNotFound",
			reqBody: models.Metrics{
				ID:    "missing",
				MType: "counter",
			},
			mockService: &mockMetricsService{
				getCounterFunc: func(name string) (int64, error) {
					return 0, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "InvalidType",
			reqBody: models.Metrics{
				ID:    "test",
				MType: "unknown",
			},
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "NameLowercaseConversion",
			reqBody: models.Metrics{
				ID:    "TEMP",
				MType: "gauge",
			},
			mockService: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					assert.Equal(t, "temp", name)
					return 1.23, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedValue:  ptr(1.23),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mockService, logger)
			r := chi.NewRouter()
			r.Post("/value", mh.ValueMetricHandlerJSON)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/value", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				respBody, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				var result models.Metrics
				assert.NoError(t, json.Unmarshal(respBody, &result))
				if tt.expectedValue != nil {
					assert.NotNil(t, result.Value)
					assert.InDelta(t, *tt.expectedValue, *result.Value, 0.001)
				}
				if tt.expectedDelta != nil {
					assert.NotNil(t, result.Delta)
					assert.Equal(t, *tt.expectedDelta, *result.Delta)
				}
			}
		})
	}
}

func TestSaveMetricHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name           string
		url            string
		mockService    *mockMetricsService
		expectedStatus int
		expectSaved    func(t *testing.T)
	}{
		{
			name:           "SaveGauge_Success",
			url:            "/update/gauge/cpu/3.14159",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T) {
				name := "cpu"
				var savedValue *float64
				mock := &mockMetricsService{
					saveGaugeFunc: func(n string, v *float64) error {
						name = n
						savedValue = v
						return nil
					},
				}

				mh := NewMetricHandler(mock, logger)
				r := chi.NewRouter()
				r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

				req := httptest.NewRequest("POST", "/update/gauge/cpu/3.14159", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "cpu", name)
				assert.NotNil(t, savedValue)
				assert.InDelta(t, 3.14159, *savedValue, 0.0001)
			},
		},
		{
			name:           "SaveCounter_Success",
			url:            "/update/counter/hits/42",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T) {
				name := "hits"
				var savedValue *int64
				mock := &mockMetricsService{
					saveCounterFunc: func(n string, v *int64) error {
						name = n
						savedValue = v
						return nil
					},
				}

				mh := NewMetricHandler(mock, logger)
				r := chi.NewRouter()
				r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

				req := httptest.NewRequest("POST", "/update/counter/hits/42", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "hits", name)
				assert.NotNil(t, savedValue)
				assert.Equal(t, int64(42), *savedValue)
			},
		},
		{
			name: "InvalidType",
			url:  "/update/unknown/cpu/1.0",
			mockService: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					t.Error("saveGaugeFunc should not be called")
					return nil
				},
				saveCounterFunc: func(name string, value *int64) error {
					t.Error("saveCounterFunc should not be called")
					return nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectSaved:    func(t *testing.T) {},
		},
		{
			name: "InvalidGaugeValue",
			url:  "/update/gauge/temp/not-a-number",
			mockService: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					t.Error("saveGaugeFunc should not be called")
					return nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectSaved:    func(t *testing.T) {},
		},
		{
			name: "InvalidCounterValue",
			url:  "/update/counter/hits/notint",
			mockService: &mockMetricsService{
				saveCounterFunc: func(name string, value *int64) error {
					t.Error("saveCounterFunc should not be called")
					return nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectSaved:    func(t *testing.T) {},
		},
		{
			name: "ServiceError",
			url:  "/update/gauge/cpu/1.23",
			mockService: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					assert.Equal(t, "cpu", name)
					assert.InDelta(t, 1.23, *value, 0.001)
					return errors.New("service error")
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectSaved:    func(t *testing.T) {},
		},
		{
			name:           "LowercaseConversion",
			url:            "/update/gauge/CPU/1.23",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T) {
				var savedName string
				mock := &mockMetricsService{
					saveGaugeFunc: func(name string, value *float64) error {
						savedName = name
						return nil
					},
				}

				mh := NewMetricHandler(mock, logger)
				r := chi.NewRouter()
				r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

				req := httptest.NewRequest("POST", "/update/gauge/CPU/1.23", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "cpu", savedName)
			},
		},
		{
			name:           "NegativeGaugeValue",
			url:            "/update/gauge/temp/-5.5",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T) {
				var savedValue *float64
				mock := &mockMetricsService{
					saveGaugeFunc: func(name string, value *float64) error {
						savedValue = value
						return nil
					},
				}

				mh := NewMetricHandler(mock, logger)
				r := chi.NewRouter()
				r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

				req := httptest.NewRequest("POST", "/update/gauge/temp/-5.5", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.NotNil(t, savedValue)
				assert.InDelta(t, -5.5, *savedValue, 0.001)
			},
		},
		{
			name:           "ZeroCounterValue",
			url:            "/update/counter/hits/0",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusOK,
			expectSaved: func(t *testing.T) {
				var savedValue *int64
				mock := &mockMetricsService{
					saveCounterFunc: func(name string, value *int64) error {
						savedValue = value
						return nil
					},
				}

				mh := NewMetricHandler(mock, logger)
				r := chi.NewRouter()
				r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

				req := httptest.NewRequest("POST", "/update/counter/hits/0", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.NotNil(t, savedValue)
				assert.Equal(t, int64(0), *savedValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mockService, logger)
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

			req := httptest.NewRequest("POST", tt.url, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			tt.expectSaved(t)
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name           string
		url            string
		mockService    *mockMetricsService
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "GetGauge_Success",
			url:  "/value/gauge/temp",
			mockService: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					assert.Equal(t, "temp", name)
					return 25.75, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "25.75",
		},
		{
			name: "GetCounter_Success",
			url:  "/value/counter/hits",
			mockService: &mockMetricsService{
				getCounterFunc: func(name string) (int64, error) {
					assert.Equal(t, "hits", name)
					return 42, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name: "GaugeNotFound",
			url:  "/value/gauge/missing",
			mockService: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					return 0, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "CounterNotFound",
			url:  "/value/counter/missing",
			mockService: &mockMetricsService{
				getCounterFunc: func(name string) (int64, error) {
					return 0, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "InvalidType",
			url:            "/value/unknown/x",
			mockService:    &mockMetricsService{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mockService, logger)
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", mh.GetMetricHandler)

			req := httptest.NewRequest("GET", tt.url, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestCollectMetricsHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name           string
		mockService    *mockMetricsService
		expectedStatus int
		expectBody     []string
	}{
		{
			name: "Success_WithMetrics",
			mockService: &mockMetricsService{
				listAllFunc: func() (map[string]float64, map[string]int64) {
					return map[string]float64{"cpu": 0.75, "memory": 0.5}, map[string]int64{"requests": 1000, "errors": 5}
				},
			},
			expectedStatus: http.StatusOK,
			expectBody:     []string{"Metrics", "cpu", "memory", "requests", "errors"},
		},
		{
			name: "EmptyMetrics",
			mockService: &mockMetricsService{
				listAllFunc: func() (map[string]float64, map[string]int64) {
					return map[string]float64{}, map[string]int64{}
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mockService, logger)
			r := chi.NewRouter()
			r.Get("/", mh.CollectMetricsHandler)

			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			for _, substr := range tt.expectBody {
				assert.Contains(t, bodyStr, substr)
			}
		})
	}
}
