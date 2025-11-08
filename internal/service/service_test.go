package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockMetricRepository struct {
	mock.Mock
}

func (m *MockMetricRepository) SaveGauge(name string, value *float64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockMetricRepository) SaveCounter(name string, value *int64) error {
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

func (m *MockMetricRepository) ListGauges() (map[string]float64, error) {
	args := m.Called()
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockMetricRepository) ListCounters() (map[string]int64, error) {
	args := m.Called()
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockMetricRepository) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	args := m.Called(gauges, counters)
	return args.Error(0)
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
			name:   "valid gauge save",
			mName:  "cpu_usage",
			mValue: 0.75,
			wantErr: false,
		},
		{
			name:      "repo error on main storage",
			mName:     "error_metric",
			mValue:    1.0,
			repoError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			logger := zap.NewNop().Sugar()
			value := tt.mValue

			mockRepo.On("SaveGauge", tt.mName, &value).Return(tt.repoError)

			service := NewMetricsService(mockRepo, mockRepo, logger)
			err := service.SaveGaugeMetric(tt.mName, &value)

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
			name:   "valid counter save",
			mName:  "requests",
			mValue: 42,
			wantErr: false,
		},
		{
			name:      "repo error on main storage",
			mName:     "error_counter",
			mValue:    100,
			repoError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			logger := zap.NewNop().Sugar()
			value := tt.mValue

			mockRepo.On("SaveCounter", tt.mName, &value).Return(tt.repoError)

			service := NewMetricsService(mockRepo, mockRepo, logger)
			err := service.SaveCounterMetric(tt.mName, &value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetGaugeMetric(t *testing.T) {
	tests := []struct {
		name      string
		mName     string
		returnVal float64
		returnErr error
		wantErr   bool
	}{
		{
			name:      "gauge found",
			mName:     "temperature",
			returnVal: 23.5,
			wantErr:   false,
		},
		{
			name:      "gauge not found",
			mName:     "missing",
			returnErr: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			logger := zap.NewNop().Sugar()

			mockRepo.On("GetGauge", tt.mName).Return(tt.returnVal, tt.returnErr)

			service := NewMetricsService(mockRepo, mockRepo, logger)
			val, err := service.GetGaugeMetric(tt.mName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, float64(0), val)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.returnVal, val)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetCounterMetric(t *testing.T) {
	tests := []struct {
		name      string
		mName     string
		returnVal int64
		returnErr error
		wantErr   bool
	}{
		{
			name:      "counter found",
			mName:     "hits",
			returnVal: 1000,
			wantErr:   false,
		},
		{
			name:      "counter not found",
			mName:     "missing",
			returnErr: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			logger := zap.NewNop().Sugar()

			mockRepo.On("GetCounter", tt.mName).Return(tt.returnVal, tt.returnErr)

			service := NewMetricsService(mockRepo, mockRepo, logger)
			val, err := service.GetCounterMetric(tt.mName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), val)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.returnVal, val)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListAllMetrics(t *testing.T) {
	tests := []struct {
		name           string
		gauges         map[string]float64
		counters       map[string]int64
		gaugesErr      error
		countersErr    error
		expectedGauges map[string]float64
		expectedCounters map[string]int64
	}{
		{
			name:           "both lists retrieved successfully",
			gauges:         map[string]float64{"cpu": 0.5},
			counters:       map[string]int64{"req": 100},
			expectedGauges: map[string]float64{"cpu": 0.5},
			expectedCounters: map[string]int64{"req": 100},
		},
		{
			name:           "error in gauges",
			gauges:         nil,
			counters:       map[string]int64{"req": 100},
			gaugesErr:      assert.AnError,
			expectedGauges: map[string]float64{},
			expectedCounters: map[string]int64{"req": 100},
		},
		{
			name:           "error in counters",
			gauges:         map[string]float64{"cpu": 0.5},
			counters:       nil,
			countersErr:    assert.AnError,
			expectedGauges: map[string]float64{"cpu": 0.5},
			expectedCounters: map[string]int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			logger := zap.NewNop().Sugar()

			mockRepo.On("ListGauges").Return(tt.gauges, tt.gaugesErr)
			mockRepo.On("ListCounters").Return(tt.counters, tt.countersErr)

			service := NewMetricsService(mockRepo, mockRepo, logger)
			gauges, counters := service.ListAllMetrics()

			assert.Equal(t, tt.expectedGauges, gauges)
			assert.Equal(t, tt.expectedCounters, counters)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSaveAllMetricsToFile(t *testing.T) {
	tests := []struct {
		name       string
		gauges     map[string]float64
		counters   map[string]int64
		repoError  error
		wantErr    bool
	}{
		{
			name:     "save all metrics successfully",
			gauges:   map[string]float64{"g1": 1.1},
			counters: map[string]int64{"c1": 2},
			wantErr:  false,
		},
		{
			name:      "repo error on save",
			gauges:    map[string]float64{"g1": 1.1},
			counters:  map[string]int64{"c1": 2},
			repoError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			logger := zap.NewNop().Sugar()

			mockRepo.On("SaveAllMetrics", tt.gauges, tt.counters).Return(tt.repoError)

			service := NewMetricsService(mockRepo, mockRepo, logger)
			err := service.SaveAllMetricsToFile(tt.gauges, tt.counters)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func ptr[T any](v T) *T { return &v }
