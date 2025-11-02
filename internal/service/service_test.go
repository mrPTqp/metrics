package service

import (
	"testing"

	"github.com/mrPTqp/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockMetricRepository - mock implementation of MetricRepository
type MockMetricRepository struct {
	mock.Mock
}

func (m *MockMetricRepository) AddGauge(name string, value *float64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockMetricRepository) AddCounter(name string, value *int64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockMetricRepository) GetGauge(name string) (float64, error) {
	args := m.Called(name)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricRepository) GetCounter(name string) (int64, error) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricRepository) ListGauges() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func (m *MockMetricRepository) ListCounters() map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]int64)
}

func TestSaveGaugeMetric(t *testing.T) {
	tests := []struct {
		name      string
		mName     string
		mValue    float64
		repoError error
		wantErr   bool
	}{
		{
			name:      "Success case",
			mName:     "test_metric",
			mValue:    123.45,
			repoError: nil,
			wantErr:   false,
		},
		{
			name:      "Repository error",
			mName:     "error_metric",
			mValue:    67.89,
			repoError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			expectedValue := tt.mValue
			mockRepo.On("AddGauge", tt.mName, &expectedValue).Return(tt.repoError)

			service := NewMetricsService(mockRepo, zap.NewNop().Sugar())
			err := service.SaveGaugeMetric(tt.mName, &tt.mValue) // передаём указатель

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSaveCounterMetric(t *testing.T) {
	tests := []struct {
		name      string
		mName     string
		mValue    int64
		repoError error
		wantErr   bool
	}{
		{
			name:      "Success case",
			mName:     "counter_metric",
			mValue:    42,
			repoError: nil,
			wantErr:   false,
		},
		{
			name:      "Repository error",
			mName:     "counter_error",
			mValue:    100,
			repoError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			expectedValue := tt.mValue
			mockRepo.On("AddCounter", tt.mName, &expectedValue).Return(tt.repoError)

			service := NewMetricsService(mockRepo, zap.NewNop().Sugar())
			err := service.SaveCounterMetric(tt.mName, &tt.mValue) // передаём указатель

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListAllMetrics(t *testing.T) {
	logger := zap.NewNop().Sugar()
	storage := repository.NewMemStorage(logger)
	metricService := NewMetricsService(storage, logger)

	_ = metricService.SaveGaugeMetric("test_gauge", ptr(3.14))
	_ = metricService.SaveCounterMetric("test_counter", ptr(int64(3)))
	_ = metricService.SaveCounterMetric("test_counter", ptr(int64(4))) // adds to existing

	gauges, counters := metricService.ListAllMetrics()

	gaugeVal, ok := gauges["test_gauge"]
	assert.True(t, ok, "expected gauge 'test_gauge' to exist")
	assert.Equal(t, 3.14, gaugeVal)

	counterVal, ok := counters["test_counter"]
	assert.True(t, ok, "expected counter 'test_counter' to exist")
	assert.Equal(t, int64(7), counterVal)
}

func ptr[T any](v T) *T { return &v }