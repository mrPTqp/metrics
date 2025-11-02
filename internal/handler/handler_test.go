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

type mockService struct {
	saveGaugeErr   error
	saveCounterErr error

	getGaugeVal float64
	getGaugeErr error
	getCountVal int64
	getCountErr error

	listGauges   map[string]float64
	listCounters map[string]int64
}

func (m *mockService) SaveGaugeMetric(name string, value *float64) error { return m.saveGaugeErr }
func (m *mockService) SaveCounterMetric(name string, value *int64) error { return m.saveCounterErr }
func (m *mockService) GetGaugeMetric(name string) (float64, error) {
	return m.getGaugeVal, m.getGaugeErr
}
func (m *mockService) GetCounterMetric(name string) (int64, error) {
	return m.getCountVal, m.getCountErr
}
func (m *mockService) ListAllMetrics() (map[string]float64, map[string]int64) {
	if m.listGauges == nil {
		m.listGauges = map[string]float64{}
	}
	if m.listCounters == nil {
		m.listCounters = map[string]int64{}
	}
	return m.listGauges, m.listCounters
}

func TestSaveMetricHandler_JSON(t *testing.T) {
	tests := []struct {
		name       string
		reqBody    models.Metrics
		mock       *mockService
		wantStatus int
	}{
		{
			name: "save gauge success",
			reqBody: models.Metrics{
				ID:    "cpu",
				MType: "gauge",
				Value: ptr(1.23),
			},
			mock:       &mockService{},
			wantStatus: http.StatusOK,
		},
		{
			name: "save counter success",
			reqBody: models.Metrics{
				ID:    "hits",
				MType: "counter",
				Delta: ptr(int64(10)),
			},
			mock:       &mockService{},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid metric type",
			reqBody: models.Metrics{
				ID:    "x",
				MType: "unknown",
			},
			mock:       &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing value for gauge",
			reqBody: models.Metrics{
				ID:    "cpu",
				MType: "gauge",
				// Value is nil
			},
			mock:       &mockService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing delta for counter",
			reqBody: models.Metrics{
				ID:    "hits",
				MType: "counter",
				// Delta is nil
			},
			mock:       &mockService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mock, zap.NewNop().Sugar())
			r := http.NewServeMux()
			r.HandleFunc("POST /update", mh.SaveMetricHandlerJSON)

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
		mock       *mockService
		wantStatus int
		wantValue  interface{} // float64 or int64
	}{
		{
			name: "get gauge success",
			reqBody: models.Metrics{
				ID:    "temp",
				MType: "gauge",
			},
			mock: &mockService{
				getGaugeVal: 3.5,
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
			mock: &mockService{
				getCountVal: 42,
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
			mock: &mockService{
				getGaugeErr: assert.AnError,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid type",
			reqBody: models.Metrics{
				ID:    "x",
				MType: "unknown",
			},
			mock:       &mockService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMetricHandler(tt.mock, zap.NewNop().Sugar())
			r := http.NewServeMux()
			r.HandleFunc("POST /value", mh.ValueMetricHandlerJSON)

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
					assert.Equal(t, tt.wantValue, *v)
				}
				if v := resp.Delta; v != nil {
					assert.Equal(t, tt.wantValue, *v)
				}
			}
		})
	}
}

// Вспомогательная функция для создания указателя
func ptr[T any](v T) *T { return &v }

func TestCollectMetricsHandler(t *testing.T) {
	mock := &mockService{
		listGauges: map[string]float64{
			"temp": 22.5,
		},
		listCounters: map[string]int64{
			"hits": 7,
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

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want %q", ct, "text/html; charset=utf-8")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Metrics") || !strings.Contains(body, "temp") || !strings.Contains(body, "hits") {
		t.Fatalf("unexpected html body: %q", body)
	}
}

func TestSaveMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		mock           *mockService
		wantStatus     int
		wantCT         string
		wantBodySubstr string
	}{
		{
			name:       "ok gauge",
			method:     http.MethodPost,
			url:        "/update/gauge/cpu/1.23",
			mock:       &mockService{},
			wantStatus: http.StatusOK,
			wantCT:     "text/plain; charset=utf-8",
		},
		{
			name:           "bad type",
			method:         http.MethodPost,
			url:            "/update/unknown/cpu/1.0",
			mock:           &mockService{},
			wantStatus:     http.StatusBadRequest,
			wantCT:         "text/plain; charset=utf-8",
			wantBodySubstr: "Invalid metric type",
		},
		{
			name:           "bad gauge value",
			method:         http.MethodPost,
			url:            "/update/gauge/temp/not-a-number",
			mock:           &mockService{},
			wantStatus:     http.StatusBadRequest,
			wantCT:         "text/plain; charset=utf-8",
			wantBodySubstr: "Invalid gauge value",
		},
		{
			name:       "ok counter",
			method:     http.MethodPost,
			url:        "/update/counter/hits/10",
			mock:       &mockService{},
			wantStatus: http.StatusOK,
			wantCT:     "text/plain; charset=utf-8",
		},
		{
			name:           "bad counter value",
			method:         http.MethodPost,
			url:            "/update/counter/hits/notint",
			mock:           &mockService{},
			wantStatus:     http.StatusBadRequest,
			wantCT:         "text/plain; charset=utf-8",
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

			if res.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tc.wantStatus)
			}
			if ct := res.Header.Get("Content-Type"); ct != tc.wantCT {
				t.Fatalf("content-type = %q, want %q", ct, tc.wantCT)
			}
			if tc.wantBodySubstr != "" {
				body := rec.Body.String()
				if !strings.Contains(body, tc.wantBodySubstr) {
					t.Fatalf("body %q does not contain %q", body, tc.wantBodySubstr)
				}
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		mock           *mockService
		wantStatus     int
		wantCT         string
		wantBodySubstr string
	}{
		{
			name:           "get gauge ok",
			method:         http.MethodGet,
			url:            "/value/gauge/temp",
			mock:           &mockService{getGaugeVal: 3.5},
			wantStatus:     http.StatusOK,
			wantCT:         "text/plain; charset=utf-8",
			wantBodySubstr: "3.5",
		},
		{
			name:           "get counter ok",
			method:         http.MethodGet,
			url:            "/value/counter/hits",
			mock:           &mockService{getCountVal: 42},
			wantStatus:     http.StatusOK,
			wantCT:         "text/plain; charset=utf-8",
			wantBodySubstr: "42",
		},
		{
			name:           "bad type",
			method:         http.MethodGet,
			url:            "/value/unknown/x",
			mock:           &mockService{},
			wantStatus:     http.StatusBadRequest,
			wantCT:         "text/plain; charset=utf-8",
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

			if res.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tc.wantStatus)
			}
			if ct := res.Header.Get("Content-Type"); ct != tc.wantCT {
				t.Fatalf("content-type = %q, want %q", ct, tc.wantCT)
			}
			if tc.wantBodySubstr != "" {
				body := rec.Body.String()
				if !strings.Contains(body, tc.wantBodySubstr) {
					t.Fatalf("body %q does not contain %q", body, tc.wantBodySubstr)
				}
			}
		})
	}
}
