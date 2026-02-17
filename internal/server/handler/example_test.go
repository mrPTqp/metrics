// internal/server/handler/example_test.go
package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ExampleMetricHandler_SaveMetricHandlerJSON() {
	logger := zap.NewNop()
	defer func() { _ = logger.Sync() }()

	mockSvc := &mockMetricsService{
		saveGaugeFunc: func(ctx context.Context, name string, value *float64) error {
			logger.Info("saving gauge", zap.String("name", name), zap.Float64("value", *value))
			return nil
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	reqBody := `{"id":"cpu_temp","type":"gauge","value":45.3}`
	req := httptest.NewRequest("POST", "/update/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SaveMetricHandlerJSON(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// {"id":"cpu_temp","type":"gauge","value":45.3}
}

func ExampleMetricHandler_SaveMetricsHandlerJSON() {
	logger := zap.NewNop()
	defer func() { _ = logger.Sync() }()

	mockSvc := &mockMetricsService{
		saveAllMetricsFunc: func(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
			logger.Info("saving multiple metrics", zap.Any("gauges", gauges), zap.Any("counters", counters))
			return nil
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	reqBody := `[{"id":"cpu","type":"gauge","value":0.8},{"id":"requests","type":"counter","delta":100}]`
	req := httptest.NewRequest("POST", "/updates/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SaveMetricsHandlerJSON(w, req)

	// Ожидается пустой ответ
	fmt.Print(w.Body.String())
	// Output:
	//
}

func ExampleMetricHandler_ValueMetricHandlerJSON() {
	logger := zap.NewNop()
	defer func() { _ = logger.Sync() }()

	mockSvc := &mockMetricsService{
		getGaugeFunc: func(ctx context.Context, name string) (float64, error) {
			if name == "cpu" {
				return 0.8, nil
			}
			return 0, fmt.Errorf("not found")
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	reqBody := `{"id":"cpu","type":"gauge"}`
	req := httptest.NewRequest("POST", "/value/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValueMetricHandlerJSON(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// {"id":"cpu","type":"gauge","value":0.8}
}

func ExampleMetricHandler_SaveMetricHandler() {
	// Создаём router, чтобы chi.URLParam работал
	r := chi.NewRouter()
	logger := zap.NewNop()
	defer func() { _ = logger.Sync() }()

	mockSvc := &mockMetricsService{
		saveGaugeFunc: func(ctx context.Context, name string, value *float64) error {
			logger.Info("saving gauge", zap.String("name", name), zap.Float64("value", *value))
			return nil
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	r.Post("/update/{type}/{name}/{value}", handler.SaveMetricHandler)

	req := httptest.NewRequest("POST", "/update/gauge/cpu_usage/0.8", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Print(w.Body.String())
	// Output:
	//
}

func ExampleMetricHandler_GetMetricHandler() {
	r := chi.NewRouter()
	logger := zap.NewNop()
	defer func() { _ = logger.Sync() }()

	mockSvc := &mockMetricsService{
		getGaugeFunc: func(ctx context.Context, name string) (float64, error) {
			if name == "cpu_usage" {
				return 0.8, nil
			}
			return 0, fmt.Errorf("not found")
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	r.Get("/value/{type}/{name}", handler.GetMetricHandler)

	req := httptest.NewRequest("GET", "/value/gauge/cpu_usage", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Print(strings.TrimSpace(w.Body.String()))
	// Output:
	// 0.8
}
