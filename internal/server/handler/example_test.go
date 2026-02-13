// internal/server/handler/example_test.go
package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func ExampleMetricHandler_SaveMetricHandlerJSON() {
	logger := zaptest.NewLogger(nil)
	defer logger.Sync()

	mockSvc := &mockMetricsService{
		saveGaugeFunc: func(ctx context.Context, name string, value *float64) error {
			logger.Info("Saving gauge", zap.String("name", name), zap.Float64("value", *value))
			return nil
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	reqBody := `{"id":"cpu_temp","mtype":"gauge","value":45.3}`
	req := httptest.NewRequest("POST", "/update/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SaveMetricHandlerJSON(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// {"id":"cpu_temp","mtype":"gauge","value":45.3}
}

func ExampleMetricHandler_SaveMetricsHandlerJSON() {
	logger := zaptest.NewLogger(nil)
	defer logger.Sync()

	mockSvc := &mockMetricsService{
		saveAllMetricsFunc: func(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
			logger.Info("Saving multiple metrics", zap.Any("gauges", gauges), zap.Any("counters", counters))
			return nil
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	reqBody := `[{"id":"cpu","mtype":"gauge","value":0.8},{"id":"requests","mtype":"counter","delta":100}]`
	req := httptest.NewRequest("POST", "/updates/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SaveMetricsHandlerJSON(w, req)

	fmt.Print(w.Body.String())
	// Output:
	//
}

func ExampleMetricHandler_ValueMetricHandlerJSON() {
	logger := zaptest.NewLogger(nil)
	defer logger.Sync()

	mockSvc := &mockMetricsService{
		getGaugeFunc: func(ctx context.Context, name string) (float64, error) {
			if name == "cpu" {
				return 0.8, nil
			}
			return 0, fmt.Errorf("not found")
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	reqBody := `{"id":"cpu","mtype":"gauge"}`
	req := httptest.NewRequest("POST", "/value/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValueMetricHandlerJSON(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// {"id":"cpu","mtype":"gauge","value":0.8}
}

func ExampleMetricHandler_SaveMetricHandler() {
	logger := zaptest.NewLogger(nil)
	defer logger.Sync()

	mockSvc := &mockMetricsService{
		saveGaugeFunc: func(ctx context.Context, name string, value *float64) error {
			logger.Info("Saving gauge", zap.String("name", name), zap.Float64("value", *value))
			return nil
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	req := httptest.NewRequest("POST", "/update/gauge/cpu_usage/0.8", nil)
	req = req.WithContext(context.WithValue(req.Context(), chiRouteCtxKey, newChiRouteContext("gauge", "cpu_usage", "0.8")))
	w := httptest.NewRecorder()

	handler.SaveMetricHandler(w, req)

	fmt.Print(w.Body.String())
	// Output:
	//
}

func ExampleMetricHandler_GetMetricHandler() {
	logger := zaptest.NewLogger(nil)
	defer logger.Sync()

	mockSvc := &mockMetricsService{
		getGaugeFunc: func(ctx context.Context, name string) (float64, error) {
			if name == "cpu_usage" {
				return 0.8, nil
			}
			return 0, fmt.Errorf("not found")
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	req := httptest.NewRequest("GET", "/value/gauge/cpu_usage", nil)
	req = req.WithContext(context.WithValue(req.Context(), chiRouteCtxKey, newChiRouteContext("gauge", "cpu_usage")))
	w := httptest.NewRecorder()

	handler.GetMetricHandler(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// 0.8
}

func ExampleMetricHandler_CollectMetricsHandler() {
	logger := zaptest.NewLogger(nil)
	defer logger.Sync()

	mockSvc := &mockMetricsService{
		listAllFunc: func(ctx context.Context) (map[string]float64, map[string]int64) {
			return map[string]float64{"cpu": 0.8}, map[string]int64{"requests": 100}
		},
	}

	handler := NewMetricHandler(mockSvc, logger)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.CollectMetricsHandler(w, req)

	fmt.Print(w.Body.String())
	// Output:
	// <html><body>
	// <h2>Gauges</h2>
	// <table border='1'>
	// <tr><td>cpu</td><td>0.8</td></tr>
	// </table>
	// <h2>Counters</h2>
	// <table border='1'>
	// <tr><td>requests</td><td>100</td></tr>
	// </table>
	// </body></html>
}

// -- Вспомогательные функции для работы с chi.RouteContext --

var chiRouteCtxKey = chi.RouteCtxKey

func newChiRouteContext(parts ...string) context.Context {
	rctx := chi.NewRouteContext()
	if len(parts) > 0 {
		rctx.URLParams.Add("type", parts[0])
	}
	if len(parts) > 1 {
		rctx.URLParams.Add("name", parts[1])
	}
	if len(parts) > 2 {
		rctx.URLParams.Add("value", parts[2])
	}
	return context.WithValue(context.Background(), chiRouteCtxKey, rctx)
}