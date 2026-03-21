package handler

import (
	"context"
	"testing"

	"github.com/mrPTqp/metrics/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockMetricWriter для тестирования
type MockMetricWriter struct {
	mock.Mock
}

func (m *MockMetricWriter) SaveGaugeMetric(ctx context.Context, name string, value *float64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockMetricWriter) SaveCounterMetric(ctx context.Context, name string, delta *int64) error {
	args := m.Called(ctx, name, delta)
	return args.Error(0)
}

func (m *MockMetricWriter) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	args := m.Called(ctx, gauges, counters)
	return args.Error(0)
}

func TestMetricsHandler_UpdateMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockWriter := &MockMetricWriter{}

	handler := &MetricsHandler{
		writer: mockWriter,
		logger: logger,
	}

	tests := []struct {
		name     string
		req      *proto.UpdateMetricsRequest
		wantErr  bool
		mockCall func()
	}{
		{
			name: "valid gauge and counter metrics",
			req: &proto.UpdateMetricsRequest{
				Metrics: []*proto.Metric{
					{
						Id:    "test_gauge",
						Type:  proto.Metric_GAUGE,
						Value: float64Ptr(1.5),
					},
					{
						Id:    "test_counter",
						Type:  proto.Metric_COUNTER,
						Delta: int64Ptr(10),
					},
				},
			},
			wantErr: false,
			mockCall: func() {
				mockWriter.On("SaveAllMetrics", mock.Anything, map[string]float64{"test_gauge": 1.5}, map[string]int64{"test_counter": 10}).Return(nil)
			},
		},
		{
			name: "empty metrics",
			req:  &proto.UpdateMetricsRequest{Metrics: []*proto.Metric{}},
			wantErr: false,
			mockCall: func() {
				mockWriter.On("SaveAllMetrics", mock.Anything, map[string]float64{}, map[string]int64{}).Return(nil)
			},
		},
		{
			name: "unknown metric type",
			req: &proto.UpdateMetricsRequest{
				Metrics: []*proto.Metric{
					{
						Id:   "test_unknown",
						Type: proto.Metric_MType(99),
					},
				},
			},
			wantErr: false,
			mockCall: func() {
				mockWriter.On("SaveAllMetrics", mock.Anything, map[string]float64{}, map[string]int64{}).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCall()
			_, err := handler.UpdateMetrics(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockWriter.AssertExpectations(t)
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}