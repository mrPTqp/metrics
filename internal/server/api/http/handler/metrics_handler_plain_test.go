package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/service"
)

type stubMetricService struct {
	savedGauges   map[string]float64
	savedCounters map[string]int64

	saveGaugeErr   error
	saveCounterErr error

	gaugeValues   map[string]float64
	counterValues map[string]int64

	saveAllGauges   map[string]float64
	saveAllCounters map[string]int64
}

func (s *stubMetricService) SaveGaugeMetric(_ context.Context, name string, value *float64) error {
	if s.savedGauges == nil {
		s.savedGauges = make(map[string]float64)
	}
	if value != nil {
		s.savedGauges[name] = *value
	}
	return s.saveGaugeErr
}

func (s *stubMetricService) SaveCounterMetric(_ context.Context, name string, value *int64) error {
	if s.savedCounters == nil {
		s.savedCounters = make(map[string]int64)
	}
	if value != nil {
		s.savedCounters[name] += *value
	}
	return s.saveCounterErr
}

func (s *stubMetricService) SaveAllMetrics(_ context.Context, gauges map[string]float64, counters map[string]int64) error {
	s.saveAllGauges = gauges
	s.saveAllCounters = counters
	return nil
}

func (s *stubMetricService) GetGaugeMetric(_ context.Context, name string) (float64, error) {
	return s.gaugeValues[name], nil
}

func (s *stubMetricService) GetCounterMetric(_ context.Context, name string) (int64, error) {
	return s.counterValues[name], nil
}

func (s *stubMetricService) ListAllMetrics(_ context.Context) (map[string]float64, map[string]int64) {
	return s.gaugeValues, s.counterValues
}

func (s *stubMetricService) Ping(_ context.Context) bool {
	return true
}

func newPlainHandler(t *testing.T, svc *stubMetricService) *MetricHandler {
	t.Helper()
	logger := zaptest.NewLogger(t)
	return &MetricHandler{
		writer: svc,
		reader: svc,
		lister: svc,
		pinger: svc,
		logger: logger,
	}
}

func newRequestWithRouteContext(method, path string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, httptest.NewRecorder()
}

func TestSaveMetricHandler_ValidationAndRouting(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValue    string
		wantStatusCode int
	}{
		{
			name:           "invalid type",
			metricType:     "unknown",
			metricName:     "m1",
			metricValue:    "1",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "empty name",
			metricType:     "gauge",
			metricName:     "",
			metricValue:    "1",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "valid gauge",
			metricType:     "gauge",
			metricName:     "g1",
			metricValue:    "1.5",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "valid counter",
			metricType:     "counter",
			metricName:     "c1",
			metricValue:    "10",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubMetricService{}
			h := newPlainHandler(t, svc)

			req, rr := newRequestWithRouteContext(http.MethodPost, "/update/{type}/{name}/{value}")
			rctx := chi.RouteContext(req.Context())
			rctx.URLParams.Add("type", tt.metricType)
			rctx.URLParams.Add("name", tt.metricName)
			rctx.URLParams.Add("value", tt.metricValue)

			req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

			h.SaveMetricHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestGetMetricHandler_ValidationAndRouting(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		wantStatusCode int
	}{
		{
			name:           "invalid type",
			metricType:     "unknown",
			metricName:     "m1",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "empty name",
			metricType:     "gauge",
			metricName:     "",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "gauge ok",
			metricType:     "gauge",
			metricName:     "g1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "counter ok",
			metricType:     "counter",
			metricName:     "c1",
			wantStatusCode: http.StatusOK,
		},
	}

	svc := &stubMetricService{
		gaugeValues:   map[string]float64{"g1": 3.14},
		counterValues: map[string]int64{"c1": 42},
	}
	h := newPlainHandler(t, svc)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, rr := newRequestWithRouteContext(http.MethodGet, "/value/{type}/{name}")
			rctx := chi.RouteContext(req.Context())
			rctx.URLParams.Add("type", tt.metricType)
			rctx.URLParams.Add("name", tt.metricName)

			req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

			h.GetMetricHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestCollectMetricsHandlerAndRenderHTML(t *testing.T) {
	svc := &stubMetricService{
		gaugeValues:   map[string]float64{"g1": 1.0},
		counterValues: map[string]int64{"c1": 2},
	}
	h := newPlainHandler(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(contextkey.WithLogger(req.Context(), zaptest.NewLogger(t)))
	rr := httptest.NewRecorder()

	h.CollectMetricsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html" {
		t.Fatalf("Content-Type = %q, want %q", ct, "text/html")
	}
	if body := rr.Body.String(); body == "" {
		t.Fatal("expected non-empty HTML body")
	}
}

func TestDBHealthCheckHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := contextkey.WithLogger(context.Background(), logger)

	tests := []struct {
		name           string
		pinger         service.Pinger
		wantStatusCode int
	}{
		{
			name:           "no pinger",
			pinger:         nil,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "unhealthy pinger",
			pinger: &struct{ service.Pinger }{
				Pinger: service.PingerFunc(func(context.Context) bool { return false }),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "healthy pinger",
			pinger: &struct{ service.Pinger }{
				Pinger: service.PingerFunc(func(context.Context) bool { return true }),
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubMetricService{}
			h := newPlainHandler(t, svc)
			h.pinger = tt.pinger

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			h.DBHealthCheckHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

