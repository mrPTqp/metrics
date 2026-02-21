package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveGauge(name string, value *float64) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *MockRepository) SaveCounter(name string, delta *int64) error {
	args := m.Called(name, delta)
	return args.Error(0)
}

func (m *MockRepository) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) {
	m.Called(gauges, counters)
}

func (m *MockRepository) SaveAdditionalGaugeMetrics(additionalGauges map[string]float64) {
	m.Called(additionalGauges)
}

func (m *MockRepository) GetAllMetrics() (map[string]float64, map[string]int64) {
	args := m.Called()
	return args.Get(0).(map[string]float64), args.Get(1).(map[string]int64)
}

func (m *MockRepository) GetAdditionalGaugeMetrics() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func TestMetricsAgent_PollMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{ReportInterval: 10, PollInterval: 2}

	tests := []struct {
		name            string
		initialCounters map[string]int64
		expectCall      func(*MockRepository)
	}{
		{
			name:            "Increments_PollCount",
			initialCounters: map[string]int64{"PollCount": 5},
			expectCall: func(mr *MockRepository) {
				mr.On("GetAllMetrics").Return(
					map[string]float64{},
					map[string]int64{"PollCount": 5},
				).Once()
				mr.On("SaveAllMetrics", mock.AnythingOfType("map[string]float64"), map[string]int64{"PollCount": 6}).
					Return().
					Once()
			},
		},
		{
			name:            "Empty_Counters_Initiates_PollCount",
			initialCounters: map[string]int64{},
			expectCall: func(mr *MockRepository) {
				mr.On("GetAllMetrics").Return(
					map[string]float64{},
					map[string]int64{},
				).Once()
				mr.On("SaveAllMetrics", mock.AnythingOfType("map[string]float64"), map[string]int64{"PollCount": 1}).
					Return().
					Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.expectCall(mockRepo)

			agent := NewMetricsAgent(&http.Client{}, cfg, mockRepo, logger)
			agent.PollMetrics()

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricsAgent_SendMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Address: models.NetAddress{Host: "localhost", Port: 8080},
	}

	tests := []struct {
		name              string
		gauges            map[string]float64
		counters          map[string]int64
		additionalGauges  map[string]float64
		serverStatus      int
		expectRequestBody bool
	}{
		{
			name:              "Sends_Valid_Metrics",
			gauges:            map[string]float64{"CPU": 0.75},
			counters:          map[string]int64{"PollCount": 42},
			additionalGauges:  map[string]float64{"RandomValue": 0.5},
			serverStatus:      http.StatusOK,
			expectRequestBody: true,
		},
		{
			name:              "Empty_Metrics_Skips_Request",
			gauges:            map[string]float64{},
			counters:          map[string]int64{},
			additionalGauges:  map[string]float64{},
			serverStatus:      http.StatusOK,
			expectRequestBody: false,
		},
		{
			name:              "Server_Returns_Error",
			gauges:            map[string]float64{"CPU": 1.0},
			counters:          map[string]int64{"Hits": 100},
			additionalGauges:  map[string]float64{},
			serverStatus:      http.StatusInternalServerError,
			expectRequestBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockRepo.On("GetAllMetrics").Return(tt.gauges, tt.counters).Once()
			mockRepo.On("GetAdditionalGaugeMetrics").Return(tt.additionalGauges).Maybe()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectRequestBody {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Greater(t, len(body), 0, "Expected non-empty request body")

					decBody, err := Decompress(body)
					assert.NoError(t, err)

					var metrics []models.Metrics
					err = json.Unmarshal(decBody, &metrics)
					assert.NoError(t, err)
					assert.True(t, len(metrics) > 0, "Expected non-empty metrics array")
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer srv.Close()

			_ = cfg.Address.SetAddress(srv.URL[len("http://"):])

			client := &http.Client{Timeout: time.Second * 5}

			agent := NewMetricsAgent(client, cfg, mockRepo, logger)
			agent.SendMetrics()

			mockRepo.AssertExpectations(t)
		})
	}
}
