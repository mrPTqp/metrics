package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type MockMetricRepository struct {
	mock.Mock
}

func (m *MockMetricRepository) SaveGauge(ctx context.Context, name string, value *float64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockMetricRepository) SaveCounter(ctx context.Context, name string, value *int64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockMetricRepository) GetGauge(ctx context.Context, name string) (float64, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricRepository) GetCounter(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricRepository) ListGauges(ctx context.Context) (map[string]float64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockMetricRepository) ListCounters(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockMetricRepository) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	args := m.Called(ctx, gauges, counters)
	return args.Error(0)
}

func (m *MockMetricRepository) CheckStorageAvailability(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockMetricRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) SaveGaugeMetric(ctx context.Context, name string, value *float64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockMetricsService) SaveCounterMetric(ctx context.Context, name string, value *int64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockMetricsService) GetGaugeMetric(ctx context.Context, name string) (float64, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricsService) GetCounterMetric(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricsService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	args := m.Called(ctx)
	var gauges map[string]float64
	var counters map[string]int64
	if args.Get(0) != nil {
		gauges = args.Get(0).(map[string]float64)
	}
	if args.Get(1) != nil {
		counters = args.Get(1).(map[string]int64)
	}
	return gauges, counters
}

func (m *MockMetricsService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	args := m.Called(ctx, gauges, counters)
	return args.Error(0)
}

func (m *MockMetricsService) Ping(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func ctxWithLogger() (context.Context, *zap.Logger) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	ctx := context.WithValue(context.Background(), zap.NewNop(), logger)
	return ctx, logger
}

func TestBaseMetricService_SaveGaugeMetric(t *testing.T) {
	testCases := []struct {
		name          string
		gaugeName     string
		value         float64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success case",
			gaugeName:     "test_gauge",
			value:         3.14159,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Repository error",
			gaugeName:     "error_gauge",
			value:         2.71828,
			mockError:     errors.New("repository error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			value := tc.value
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("SaveGauge", ctx, tc.gaugeName, &value).Return(tc.mockError)

			service := NewMetricsService(mockRepo, logger)
			err := service.SaveGaugeMetric(ctx, tc.gaugeName, &value)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBaseMetricService_SaveCounterMetric(t *testing.T) {
	testCases := []struct {
		name          string
		counterName   string
		value         int64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success case",
			counterName:   "test_counter",
			value:         42,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Repository error",
			counterName:   "error_counter",
			value:         100,
			mockError:     errors.New("repository error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			value := tc.value
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("SaveCounter", ctx, tc.counterName, &value).Return(tc.mockError)

			service := NewMetricsService(mockRepo, logger)
			err := service.SaveCounterMetric(ctx, tc.counterName, &value)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBaseMetricService_GetGaugeMetric(t *testing.T) {
	testCases := []struct {
		name          string
		gaugeName     string
		returnValue   float64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success case",
			gaugeName:     "test_gauge",
			returnValue:   25.75,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Not found",
			gaugeName:     "missing_gauge",
			returnValue:   0,
			mockError:     errors.New("not found"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("GetGauge", ctx, tc.gaugeName).Return(tc.returnValue, tc.mockError)

			service := NewMetricsService(mockRepo, logger)
			value, err := service.GetGaugeMetric(ctx, tc.gaugeName)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
				assert.Equal(t, tc.returnValue, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.returnValue, value)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBaseMetricService_GetCounterMetric(t *testing.T) {
	testCases := []struct {
		name          string
		counterName   string
		returnValue   int64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success case",
			counterName:   "test_counter",
			returnValue:   500,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Not found",
			counterName:   "missing_counter",
			returnValue:   0,
			mockError:     errors.New("not found"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("GetCounter", ctx, tc.counterName).Return(tc.returnValue, tc.mockError)

			service := NewMetricsService(mockRepo, logger)
			value, err := service.GetCounterMetric(ctx, tc.counterName)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
				assert.Equal(t, tc.returnValue, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.returnValue, value)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBaseMetricService_ListAllMetrics(t *testing.T) {
	testCases := []struct {
		name             string
		gaugesReturn     map[string]float64
		countersReturn   map[string]int64
		gaugesErr        error
		countersErr      error
		expectedGauges   map[string]float64
		expectedCounters map[string]int64
	}{
		{
			name:             "Success case",
			gaugesReturn:     map[string]float64{"cpu": 0.75, "memory": 0.50},
			countersReturn:   map[string]int64{"requests": 1000, "errors": 5},
			gaugesErr:        nil,
			countersErr:      nil,
			expectedGauges:   map[string]float64{"cpu": 0.75, "memory": 0.50},
			expectedCounters: map[string]int64{"requests": 1000, "errors": 5},
		},
		{
			name:             "Partial error - counters error",
			gaugesReturn:     map[string]float64{"cpu": 0.75},
			countersReturn:   nil,
			gaugesErr:        nil,
			countersErr:      errors.New("list counters error"),
			expectedGauges:   map[string]float64{"cpu": 0.75},
			expectedCounters: map[string]int64{},
		},
		{
			name:             "Both errors",
			gaugesReturn:     nil,
			countersReturn:   nil,
			gaugesErr:        errors.New("list gauges error"),
			countersErr:      errors.New("list counters error"),
			expectedGauges:   map[string]float64{},
			expectedCounters: map[string]int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("ListGauges", ctx).Return(tc.gaugesReturn, tc.gaugesErr)
			mockRepo.On("ListCounters", ctx).Return(tc.countersReturn, tc.countersErr)

			service := NewMetricsService(mockRepo, logger)
			gauges, counters := service.ListAllMetrics(ctx)

			assert.Equal(t, tc.expectedGauges, gauges)
			assert.Equal(t, tc.expectedCounters, counters)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBaseMetricService_SaveAllMetrics(t *testing.T) {
	testCases := []struct {
		name          string
		gauges        map[string]float64
		counters      map[string]int64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success case",
			gauges:        map[string]float64{"cpu": 0.75, "memory": 0.50},
			counters:      map[string]int64{"requests": 1000, "errors": 5},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Error case",
			gauges:        map[string]float64{"cpu": 0.75},
			counters:      map[string]int64{"requests": 1000},
			mockError:     errors.New("save error"),
			expectedError: true,
		},
		{
			name:          "Empty maps",
			gauges:        make(map[string]float64),
			counters:      make(map[string]int64),
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("SaveAllMetrics", ctx, tc.gauges, tc.counters).Return(tc.mockError)

			service := NewMetricsService(mockRepo, logger)
			err := service.SaveAllMetrics(ctx, tc.gauges, tc.counters)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBaseMetricService_Ping(t *testing.T) {
	testCases := []struct {
		name           string
		availability   bool
		expectedResult bool
	}{
		{
			name:           "Available",
			availability:   true,
			expectedResult: true,
		},
		{
			name:           "Unavailable",
			availability:   false,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			mockRepo.On("CheckStorageAvailability", ctx).Return(tc.availability)

			service := NewMetricsService(mockRepo, logger)
			result := service.Ping(ctx)

			assert.Equal(t, tc.expectedResult, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFileBackupService_SaveGaugeMetric(t *testing.T) {
	testCases := []struct {
		name             string
		gaugeName        string
		value            float64
		baseServiceError error
		fileBackupError  error
		expectError      bool
		expectedError    error
		expectFileBackup bool
	}{
		{
			name:             "Success",
			gaugeName:        "cpu",
			value:            1.23,
			baseServiceError: nil,
			fileBackupError:  nil,
			expectError:      false,
			expectFileBackup: true,
		},
		{
			name:             "Base service error",
			gaugeName:        "cpu",
			value:            1.23,
			baseServiceError: errors.New("base service error"),
			fileBackupError:  nil,
			expectError:      true,
			expectedError:    errors.New("base service error"),
			expectFileBackup: false,
		},
		{
			name:             "File backup error",
			gaugeName:        "cpu",
			value:            1.23,
			baseServiceError: nil,
			fileBackupError:  errors.New("file backup error"),
			expectError:      false,
			expectFileBackup: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseService := new(MockMetricsService)
			fileRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)
			value := tc.value

			baseService.On("SaveGaugeMetric", ctx, tc.gaugeName, &value).Return(tc.baseServiceError)
			if tc.expectFileBackup {
				fileRepo.On("SaveGauge", ctx, tc.gaugeName, &value).Return(tc.fileBackupError)
			}

			decorator := NewFileBackupService(baseService, fileRepo, logger)
			err := decorator.SaveGaugeMetric(ctx, tc.gaugeName, &value)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			baseService.AssertExpectations(t)
			if tc.expectFileBackup {
				fileRepo.AssertExpectations(t)
			} else {
				fileRepo.AssertNotCalled(t, "SaveGauge")
			}
		})
	}
}

func TestFileBackupService_SaveCounterMetric(t *testing.T) {
	testCases := []struct {
		name             string
		counterName      string
		value            int64
		baseServiceError error
		fileBackupError  error
		expectError      bool
		expectedError    error
		expectFileBackup bool
	}{
		{
			name:             "Success",
			counterName:      "hits",
			value:            50,
			baseServiceError: nil,
			fileBackupError:  nil,
			expectError:      false,
			expectFileBackup: true,
		},
		{
			name:             "Base service error",
			counterName:      "hits",
			value:            50,
			baseServiceError: errors.New("base service error"),
			fileBackupError:  nil,
			expectError:      true,
			expectedError:    errors.New("base service error"),
			expectFileBackup: false,
		},
		{
			name:             "File backup error",
			counterName:      "hits",
			value:            50,
			baseServiceError: nil,
			fileBackupError:  errors.New("file backup error"),
			expectError:      false,
			expectFileBackup: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseService := new(MockMetricsService)
			fileRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)
			value := tc.value

			baseService.On("SaveCounterMetric", ctx, tc.counterName, &value).Return(tc.baseServiceError)
			if tc.expectFileBackup {
				fileRepo.On("SaveCounter", ctx, tc.counterName, &value).Return(tc.fileBackupError)
			}

			decorator := NewFileBackupService(baseService, fileRepo, logger)
			err := decorator.SaveCounterMetric(ctx, tc.counterName, &value)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			baseService.AssertExpectations(t)
			if tc.expectFileBackup {
				fileRepo.AssertExpectations(t)
			} else {
				fileRepo.AssertNotCalled(t, "SaveCounter")
			}
		})
	}
}

func TestFileBackupService_GetGaugeMetric(t *testing.T) {
	testCases := []struct {
		name          string
		gaugeName     string
		returnValue   float64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success",
			gaugeName:     "temp",
			returnValue:   20.5,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Error",
			gaugeName:     "missing",
			returnValue:   0,
			mockError:     errors.New("not found"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseService := new(MockMetricsService)
			fileRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			baseService.On("GetGaugeMetric", ctx, tc.gaugeName).Return(tc.returnValue, tc.mockError)

			decorator := NewFileBackupService(baseService, fileRepo, logger)
			value, err := decorator.GetGaugeMetric(ctx, tc.gaugeName)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.returnValue, value)
			}

			baseService.AssertExpectations(t)
		})
	}
}

func TestFileBackupService_GetCounterMetric(t *testing.T) {
	testCases := []struct {
		name          string
		counterName   string
		returnValue   int64
		mockError     error
		expectedError bool
	}{
		{
			name:          "Success",
			counterName:   "hits",
			returnValue:   42,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Error",
			counterName:   "missing",
			returnValue:   0,
			mockError:     errors.New("not found"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseService := new(MockMetricsService)
			fileRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			baseService.On("GetCounterMetric", ctx, tc.counterName).Return(tc.returnValue, tc.mockError)

			decorator := NewFileBackupService(baseService, fileRepo, logger)
			value, err := decorator.GetCounterMetric(ctx, tc.counterName)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.returnValue, value)
			}

			baseService.AssertExpectations(t)
		})
	}
}

func TestFileBackupService_ListAllMetrics(t *testing.T) {
	testCases := []struct {
		name     string
		gauges   map[string]float64
		counters map[string]int64
	}{
		{
			name:     "Success",
			gauges:   map[string]float64{"cpu": 0.8},
			counters: map[string]int64{"req": 1000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseService := new(MockMetricsService)
			fileRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			baseService.On("ListAllMetrics", ctx).Return(tc.gauges, tc.counters)

			decorator := NewFileBackupService(baseService, fileRepo, logger)
			gauges, counters := decorator.ListAllMetrics(ctx)

			assert.Equal(t, tc.gauges, gauges)
			assert.Equal(t, tc.counters, counters)

			baseService.AssertExpectations(t)
		})
	}
}

func TestFileBackupService_SaveAllMetrics(t *testing.T) {
	testCases := []struct {
		name             string
		gauges           map[string]float64
		counters         map[string]int64
		baseServiceError error
		fileBackupError  error
		expectError      bool
		expectedError    error
		expectFileBackup bool
	}{
		{
			name:             "Success",
			gauges:           map[string]float64{"g1": 1.1},
			counters:         map[string]int64{"c1": 2},
			baseServiceError: nil,
			fileBackupError:  nil,
			expectError:      false,
			expectFileBackup: true,
		},
		{
			name:             "Base service error",
			gauges:           map[string]float64{"g1": 1.1},
			counters:         map[string]int64{"c1": 2},
			baseServiceError: errors.New("base service error"),
			fileBackupError:  nil,
			expectError:      true,
			expectedError:    errors.New("base service error"),
			expectFileBackup: false,
		},
		{
			name:             "File backup error",
			gauges:           map[string]float64{"g1": 1.1},
			counters:         map[string]int64{"c1": 2},
			baseServiceError: nil,
			fileBackupError:  errors.New("file backup error"),
			expectError:      true,
			expectedError:    errors.New("file backup error"),
			expectFileBackup: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseService := new(MockMetricsService)
			fileRepo := new(MockMetricRepository)
			_, logger := ctxWithLogger()
			ctx := context.WithValue(context.Background(), zap.NewNop(), logger)

			baseService.On("SaveAllMetrics", ctx, tc.gauges, tc.counters).Return(tc.baseServiceError)

			if tc.expectFileBackup {
				fileRepo.On("SaveAllMetrics", ctx, tc.gauges, tc.counters).Return(tc.fileBackupError)
			}

			decorator := NewFileBackupService(baseService, fileRepo, logger)
			err := decorator.SaveAllMetrics(ctx, tc.gauges, tc.counters)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			baseService.AssertExpectations(t)
			if tc.expectFileBackup {
				fileRepo.AssertExpectations(t)
			} else {
				fileRepo.AssertNotCalled(t, "SaveAllMetrics")
			}
		})
	}
}
