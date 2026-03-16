package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/models"
)

func newJSONHandler(t *testing.T, svc *stubMetricService) *MetricHandler {
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

func TestSaveMetricHandlerJSON_ContentTypeAndValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	type testCase struct {
		name           string
		contentType    string
		body           interface{}
		wantStatusCode int
	}

	tests := []testCase{
		{
			name:           "wrong content type",
			contentType:    "text/plain",
			body:           nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid json body",
			contentType:    "application/json",
			body:           "{invalid",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "unsupported metric type",
			contentType: "application/json",
			body: models.Metrics{
				ID:    "m1",
				MType: "unknown",
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubMetricService{}
			h := newJSONHandler(t, svc)

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case nil:
				bodyBytes = nil
			case string:
				bodyBytes = []byte(b)
			default:
				var err error
				bodyBytes, err = json.Marshal(b)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)
			req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

			rr := httptest.NewRecorder()

			h.SaveMetricHandlerJSON(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestSaveMetricHandlerJSON_GaugeAndCounter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := contextkey.WithLogger(context.Background(), logger)

	gaugeValue := 3.14
	counterDelta := int64(5)

	tests := []struct {
		name           string
		metric         models.Metrics
		wantStatusCode int
	}{
		{
			name: "gauge missing value",
			metric: models.Metrics{
				ID:    "g1",
				MType: "gauge",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "counter missing delta",
			metric: models.Metrics{
				ID:    "c1",
				MType: "counter",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "gauge ok",
			metric: models.Metrics{
				ID:    "g1",
				MType: "gauge",
				Value: &gaugeValue,
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "counter ok",
			metric: models.Metrics{
				ID:    "c1",
				MType: "counter",
				Delta: &counterDelta,
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubMetricService{}
			h := newJSONHandler(t, svc)

			body, err := json.Marshal(tt.metric)
			if err != nil {
				t.Fatalf("marshal metric: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			h.SaveMetricHandlerJSON(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestSaveMetricsHandlerJSON_Batch(t *testing.T) {
	logger := zaptest.NewLogger(t)

	gv := 1.0
	cv := int64(2)

	tests := []struct {
		name           string
		contentType    string
		body           interface{}
		wantStatusCode int
	}{
		{
			name:           "wrong content type",
			contentType:    "text/plain",
			body:           nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			contentType:    "application/json",
			body:           "{invalid",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "gauge missing value",
			contentType: "application/json",
			body: []models.Metrics{
				{ID: "g1", MType: "gauge"},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "counter missing delta",
			contentType: "application/json",
			body: []models.Metrics{
				{ID: "c1", MType: "counter"},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "ok gauges and counters",
			contentType: "application/json",
			body: []models.Metrics{
				{ID: "g1", MType: "gauge", Value: &gv},
				{ID: "c1", MType: "counter", Delta: &cv},
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubMetricService{}
			h := newJSONHandler(t, svc)

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case nil:
				bodyBytes = nil
			case string:
				bodyBytes = []byte(b)
			default:
				var err error
				bodyBytes, err = json.Marshal(b)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)
			req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

			rr := httptest.NewRecorder()

			h.SaveMetricsHandlerJSON(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestValueMetricHandlerJSON_GetGaugeAndCounter(t *testing.T) {
	logger := zaptest.NewLogger(t)

	gv := 2.5
	cv := int64(10)

	svc := &stubMetricService{
		gaugeValues:   map[string]float64{"g1": gv},
		counterValues: map[string]int64{"c1": cv},
	}
	h := newJSONHandler(t, svc)

	tests := []struct {
		name           string
		metric         models.Metrics
		wantStatusCode int
	}{
		{
			name: "wrong content type",
			metric: models.Metrics{
				ID:    "g1",
				MType: "gauge",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "unsupported type",
			metric: models.Metrics{
				ID:    "m1",
				MType: "unknown",
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "gauge ok",
			metric: models.Metrics{
				ID:    "g1",
				MType: "gauge",
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "counter ok",
			metric: models.Metrics{
				ID:    "c1",
				MType: "counter",
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.metric)
			if err != nil {
				t.Fatalf("marshal metric: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
			if tt.name != "wrong content type" {
				req.Header.Set("Content-Type", "application/json")
			}
			req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

			rr := httptest.NewRecorder()

			h.ValueMetricHandlerJSON(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

