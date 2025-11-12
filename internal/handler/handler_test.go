// internal/handler/handler_test.go
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Мок теперь имплементирует интерфейс MetricsService
type mockMetricsService struct {
	saveGaugeFunc    func(name string, value *float64) error
	saveCounterFunc  func(name string, value *int64) error
	getGaugeFunc     func(name string) (float64, error)
	getCounterFunc   func(name string) (int64, error)
	listAllFunc      func() (map[string]float64, map[string]int64)
	saveAllMetricsFunc func(gauges map[string]float64, counters map[string]int64) error
}

func (m *mockMetricsService) SaveGaugeMetric(name string, value *float64) error {
	return m.saveGaugeFunc(name, value)
}

func (m *mockMetricsService) SaveCounterMetric(name string, value *int64) error {
	return m.saveCounterFunc(name, value)
}

func (m *mockMetricsService) GetGaugeMetric(name string) (float64, error) {
	return m.getGaugeFunc(name)
}

func (m *mockMetricsService) GetCounterMetric(name string) (int64, error) {
	return m.getCounterFunc(name)
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

func TestSaveMetricHandler_JSON(t *testing.T) {
	tests := []struct {
		name       string
		reqBody    models.Metrics
		mock       *mockMetricsService
		wantStatus int
	}{
		{
			name: "save gauge success",
			reqBody: models.Metrics{
				ID:    "cpu",
				MType: "gauge",
				Value: ptr(1.23),
			},
			mock: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					assert.Equal(t, "cpu", name)
					assert.InDelta(t, 1.23, *value, 0.001)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "save counter success",
			reqBody: models.Metrics{
				ID:    "hits",
				MType: "counter",
				Delta: ptr(int64(10)),
			},
			mock: &mockMetricsService{
				saveCounterFunc: func(name string, value *int64) error {
					assert.Equal(t, "hits", name)
					assert.Equal(t, int64(10), *value)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid metric type",
			reqBody: models.Metrics{
				ID:    "x",
				MType: "unknown",
			},
			mock:       &mockMetricsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing value for gauge",
			reqBody: models.Metrics{
				ID:    "cpu",
				MType: "gauge",
			},
			mock:       &mockMetricsService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing delta for counter",
			reqBody: models.Metrics{
				ID:    "hits",
				MType: "counter",
			},
			mock:       &mockMetricsService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mock, zap.NewNop().Sugar())
			r := chi.NewRouter()
			r.Post("/update", mh.SaveMetricHandlerJSON)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestValueMetricHandler_JSON(t *testing.T) {
	tests := []struct {
		name       string
		reqBody    models.Metrics
		mock       *mockMetricsService
		wantStatus int
		wantValue  interface{}
	}{
		{
			name: "get gauge success",
			reqBody: models.Metrics{
				ID:    "temp",
				MType: "gauge",
			},
			mock: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					assert.Equal(t, "temp", name)
					return 3.5, nil
				},
			},
			wantStatus: http.StatusOK,
			wantValue:  3.5,
		},
		{
			name: "get counter success",
			reqBody: models.Metrics{
				ID:    "hits",
				MType: "counter",
			},
			mock: &mockMetricsService{
				getCounterFunc: func(name string) (int64, error) {
					assert.Equal(t, "hits", name)
					return 42, nil
				},
			},
			wantStatus: http.StatusOK,
			wantValue:  int64(42),
		},
		{
			name: "metric not found",
			reqBody: models.Metrics{
				ID:    "missing",
				MType: "gauge",
			},
			mock: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					return 0, assert.AnError
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid type",
			reqBody: models.Metrics{
				ID:    "x",
				MType: "unknown",
			},
			mock:       &mockMetricsService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mock, zap.NewNop().Sugar())
			r := chi.NewRouter()
			r.Post("/value", mh.ValueMetricHandlerJSON)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/value", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var resp models.Metrics
				_ = json.Unmarshal(rec.Body.Bytes(), &resp)
				if v := resp.Value; v != nil {
					assert.InDelta(t, tt.wantValue.(float64), *v, 0.001)
				}
				if v := resp.Delta; v != nil {
					assert.Equal(t, tt.wantValue.(int64), *v)
				}
			}
		})
	}
}

func TestCollectMetricsHandler(t *testing.T) {
	mock := &mockMetricsService{
		listAllFunc: func() (map[string]float64, map[string]int64) {
			return map[string]float64{"temp": 22.5}, map[string]int64{"hits": 7}
		},
	}
	mh := NewMetricHandler(mock, zap.NewNop().Sugar())
	r := chi.NewRouter()
	r.Get("/", mh.CollectMetricsHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() { _ = res.Body.Close() }()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/html", res.Header.Get("Content-Type"))

	body := rec.Body.String()
	assert.Contains(t, body, "Metrics")
	assert.Contains(t, body, "temp")
	assert.Contains(t, body, "hits")
}

func TestSaveMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		mock           *mockMetricsService
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:   "ok gauge",
			method: http.MethodPost,
			url:    "/update/gauge/cpu/1.23",
			mock: &mockMetricsService{
				saveGaugeFunc: func(name string, value *float64) error {
					assert.Equal(t, "cpu", name)
					assert.InDelta(t, 1.23, *value, 0.001)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "bad type",
			method:         http.MethodPost,
			url:            "/update/unknown/cpu/1.0",
			mock:           &mockMetricsService{},
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid metric type",
		},
		{
			name:           "bad gauge value",
			method:         http.MethodPost,
			url:            "/update/gauge/temp/not-a-number",
			mock:           &mockMetricsService{},
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid gauge value",
		},
		{
			name:   "ok counter",
			method: http.MethodPost,
			url:    "/update/counter/hits/10",
			mock: &mockMetricsService{
				saveCounterFunc: func(name string, value *int64) error {
					assert.Equal(t, "hits", name)
					assert.Equal(t, int64(10), *value)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "bad counter value",
			method:         http.MethodPost,
			url:            "/update/counter/hits/notint",
			mock:           &mockMetricsService{},
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid counter value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := NewMetricHandler(tc.mock, zap.NewNop().Sugar())
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", mh.SaveMetricHandler)

			req := httptest.NewRequest(tc.method, tc.url, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			res := rec.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			assert.Equal(t, tc.wantStatus, res.StatusCode)

			if tc.wantBodySubstr != "" {
				body := rec.Body.String()
				assert.Contains(t, body, tc.wantBodySubstr)
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		mock           *mockMetricsService
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:   "get gauge ok",
			method: http.MethodGet,
			url:    "/value/gauge/temp",
			mock: &mockMetricsService{
				getGaugeFunc: func(name string) (float64, error) {
					assert.Equal(t, "temp", name)
					return 3.5, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "3.5",
		},
		{
			name:   "get counter ok",
			method: http.MethodGet,
			url:    "/value/counter/hits",
			mock: &mockMetricsService{
				getCounterFunc: func(name string) (int64, error) {
					assert.Equal(t, "hits", name)
					return 42, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "42",
		},
		{
			name:           "bad type",
			method:         http.MethodGet,
			url:            "/value/unknown/x",
			mock:           &mockMetricsService{},
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid metric type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := NewMetricHandler(tc.mock, zap.NewNop().Sugar())
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", mh.GetMetricHandler)

			req := httptest.NewRequest(tc.method, tc.url, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			res := rec.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			assert.Equal(t, tc.wantStatus, res.StatusCode)

			if tc.wantBodySubstr != "" {
				body := rec.Body.String()
				assert.Contains(t, body, tc.wantBodySubstr)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
