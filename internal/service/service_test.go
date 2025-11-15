// internal/service/service_test.go
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// --- Мок репозитория ---
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

// --- Тесты для BaseMetricService ---
func TestBaseMetricService_SaveGauge(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()
	value := 0.75

	mockRepo.On("SaveGauge", "cpu", &value).Return(nil)

	service := NewMetricsService(mockRepo, logger)
	err := service.SaveGaugeMetric("cpu", &value)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBaseMetricService_SaveGauge_Error(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()
	value := 1.0

	mockRepo.On("SaveGauge", "err", &value).Return(assert.AnError)

	service := NewMetricsService(mockRepo, logger)
	err := service.SaveGaugeMetric("err", &value)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBaseMetricService_SaveCounter(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()
	value := int64(100)

	mockRepo.On("SaveCounter", "req", &value).Return(nil)

	service := NewMetricsService(mockRepo, logger)
	err := service.SaveCounterMetric("req", &value)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBaseMetricService_GetGauge(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()

	mockRepo.On("GetGauge", "temp").Return(25.5, nil)

	service := NewMetricsService(mockRepo, logger)
	val, err := service.GetGaugeMetric("temp")

	assert.NoError(t, err)
	assert.Equal(t, 25.5, val)
	mockRepo.AssertExpectations(t)
}

func TestBaseMetricService_GetCounter(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()

	mockRepo.On("GetCounter", "hits").Return(int64(500), nil)

	service := NewMetricsService(mockRepo, logger)
	val, err := service.GetCounterMetric("hits")

	assert.NoError(t, err)
	assert.Equal(t, int64(500), val)
	mockRepo.AssertExpectations(t)
}

func TestBaseMetricService_ListAllMetrics(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()

	mockRepo.On("ListGauges").Return(map[string]float64{"cpu": 0.8}, nil)
	mockRepo.On("ListCounters").Return(map[string]int64{"req": 1000}, nil)

	service := NewMetricsService(mockRepo, logger)
	gauges, counters := service.ListAllMetrics()

	assert.Equal(t, 0.8, gauges["cpu"])
	assert.Equal(t, int64(1000), counters["req"])
	mockRepo.AssertExpectations(t)
}

func TestBaseMetricService_SaveAllMetrics(t *testing.T) {
	mockRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()

	gauges := map[string]float64{"cpu": 0.5}
	counters := map[string]int64{"req": 10}

	mockRepo.On("SaveAllMetrics", gauges, counters).Return(nil)

	service := NewMetricsService(mockRepo, logger)
	err := service.SaveAllMetrics(gauges, counters)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- Тесты для FileBackupService ---
func TestFileBackupService_SaveGauge(t *testing.T) {
	baseService := new(MockMetricsService)
	fileRepo := new(MockMetricRepository)

	logger := zap.NewNop().Sugar()
	value := 1.23

	baseService.On("SaveGaugeMetric", "cpu", &value).Return(nil)
	fileRepo.On("SaveGauge", "cpu", &value).Return(nil)

	decorator := NewFileBackupService(baseService, fileRepo, logger)
	err := decorator.SaveGaugeMetric("cpu", &value)

	assert.NoError(t, err)
	baseService.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
}

func TestFileBackupService_SaveGauge_BaseError(t *testing.T) {
	baseService := new(MockMetricsService)
	fileRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()
	value := 1.23

	baseService.On("SaveGaugeMetric", "cpu", &value).Return(assert.AnError)

	decorator := NewFileBackupService(baseService, fileRepo, logger)
	err := decorator.SaveGaugeMetric("cpu", &value)

	assert.Error(t, err)
	baseService.AssertExpectations(t)
	fileRepo.AssertNotCalled(t, "SaveGauge") // Не должен сохранять в файл, если основной сервис упал
}

func TestFileBackupService_SaveCounter(t *testing.T) {
	baseService := new(MockMetricsService)
	fileRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()
	value := int64(50)

	baseService.On("SaveCounterMetric", "hits", &value).Return(nil)
	fileRepo.On("SaveCounter", "hits", &value).Return(nil)

	decorator := NewFileBackupService(baseService, fileRepo, logger)
	err := decorator.SaveCounterMetric("hits", &value)

	assert.NoError(t, err)
	baseService.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
}

func TestFileBackupService_GetGauge(t *testing.T) {
	baseService := new(MockMetricsService)
	fileRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()

	baseService.On("GetGaugeMetric", "temp").Return(20.5, nil)

	decorator := NewFileBackupService(baseService, fileRepo, logger)
	val, err := decorator.GetGaugeMetric("temp")

	assert.NoError(t, err)
	assert.Equal(t, 20.5, val)
	baseService.AssertExpectations(t)
}

func TestFileBackupService_SaveAllMetrics(t *testing.T) {
	baseService := new(MockMetricsService)
	fileRepo := new(MockMetricRepository)
	logger := zap.NewNop().Sugar()

	gauges := map[string]float64{"g1": 1.1}
	counters := map[string]int64{"c1": 2}

	baseService.On("SaveAllMetrics", gauges, counters).Return(nil)
	fileRepo.On("SaveAllMetrics", gauges, counters).Return(nil)

	decorator := NewFileBackupService(baseService, fileRepo, logger)
	err := decorator.SaveAllMetrics(gauges, counters)

	assert.NoError(t, err)
	baseService.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
}

// --- Mock для сервиса (для тестирования декоратора) ---
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) SaveGaugeMetric(name string, value *float64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockMetricsService) SaveCounterMetric(name string, value *int64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockMetricsService) GetGaugeMetric(name string) (float64, error) {
	args := m.Called(name)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricsService) GetCounterMetric(name string) (int64, error) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricsService) ListAllMetrics() (map[string]float64, map[string]int64) {
	args := m.Called()
	return args.Get(0).(map[string]float64), args.Get(1).(map[string]int64)
}

func (m *MockMetricsService) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	args := m.Called(gauges, counters)
	return args.Error(0)
}
