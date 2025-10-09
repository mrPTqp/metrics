package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricRepository - mock implementation of MetricRepository
type MockMetricRepository struct {
	mock.Mock
}

func (m *MockMetricRepository) AddGauge(name string, value float64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockMetricRepository) AddCounter(name string, value int64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func TestProcessGaugeMetric(t *testing.T) {
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
			mockRepo.On("AddGauge", tt.mName, tt.mValue).Return(tt.repoError)

			service := NewMetricsService(mockRepo)
			err := service.ProcessGaugeMetric(tt.mName, tt.mValue)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProcessCounterMetric(t *testing.T) {
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
			mockRepo.On("AddCounter", tt.mName, tt.mValue).Return(tt.repoError)

			service := NewMetricsService(mockRepo)
			err := service.ProcessCounterMetric(tt.mName, tt.mValue)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
