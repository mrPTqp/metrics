package agent

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/proto"
	"github.com/mrPTqp/metrics/internal/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MockMetricsClient для тестирования
type MockMetricsClient struct {
	mock.Mock
}

func (m *MockMetricsClient) UpdateMetrics(ctx context.Context, in *proto.UpdateMetricsRequest, opts ...grpc.CallOption) (*proto.UpdateMetricsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*proto.UpdateMetricsResponse), args.Error(1)
}

// MockRepository для тестирования
type MockGRPCRepository struct {
	mock.Mock
}

func (m *MockGRPCRepository) GetAllMetrics() (map[string]float64, map[string]int64) {
	args := m.Called()
	return args.Get(0).(map[string]float64), args.Get(1).(map[string]int64)
}

func (m *MockGRPCRepository) GetAdditionalGaugeMetrics() map[string]float64 {
	args := m.Called()
	return args.Get(0).(map[string]float64)
}

func TestGRPCMetricsAgent_SendMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockClient := &MockMetricsClient{}
	mockRepo := &MockRepository{}

	// Создаем GRPCMetricsAgent с mock-клиентом
	agent := &GRPCMetricsAgent{
		client: mockClient,
		cfg: &config.Config{
			GRPCServerAddress: models.NetAddress{}, // Заглушка
		},
		logger: logger,
	}

	tests := []struct {
		name     string
		gauges   map[string]float64
		counters map[string]int64
		addGauges map[string]float64
		wantErr  bool
		mockCall func()
	}{
		{
			name:     "valid metrics",
			gauges:   map[string]float64{"gauge1": 1.5},
			counters: map[string]int64{"counter1": 10},
			addGauges: map[string]float64{"add_gauge1": 2.5},
			wantErr:  false,
			mockCall: func() {
				mockRepo.On("GetAllMetrics").Return(map[string]float64{"gauge1": 1.5}, map[string]int64{"counter1": 10})
				mockRepo.On("GetAdditionalGaugeMetrics").Return(map[string]float64{"add_gauge1": 2.5})
				mockClient.On("UpdateMetrics", mock.Anything, mock.AnythingOfType("*proto.UpdateMetricsRequest"), mock.Anything).Return(&proto.UpdateMetricsResponse{}, nil)
			},
		},
		{
			name:     "empty metrics",
			gauges:   map[string]float64{},
			counters: map[string]int64{},
			addGauges: map[string]float64{},
			wantErr:  false,
			mockCall: func() {
				mockRepo.On("GetAllMetrics").Return(map[string]float64{}, map[string]int64{})
				mockRepo.On("GetAdditionalGaugeMetrics").Return(map[string]float64{})
			},
		},
		{
			name:     "client error",
			gauges:   map[string]float64{"gauge1": 1.5},
			counters: map[string]int64{"counter1": 10},
			addGauges: map[string]float64{},
			wantErr:  true,
			mockCall: func() {
				mockRepo.On("GetAllMetrics").Return(map[string]float64{"gauge1": 1.5}, map[string]int64{"counter1": 10})
				mockRepo.On("GetAdditionalGaugeMetrics").Return(map[string]float64{})
				mockClient.On("UpdateMetrics", mock.Anything, mock.AnythingOfType("*proto.UpdateMetricsRequest"), mock.Anything).Return((*proto.UpdateMetricsResponse)(nil), errors.New("gRPC error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCall()
			agent.SendMetrics(context.Background(), mockRepo)
			mockClient.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetClientIP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ip := GetClientIP(logger)
	assert.NotEmpty(t, ip)
	// Проверяем, что это валидный IP
	parsed := net.ParseIP(ip)
	assert.NotNil(t, parsed)
}

func TestNewGRPCMetricsAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		GRPCServerAddress: models.NetAddress{}, // Заглушка
	}
	metadataTransport := NewGRPCMetadataTransport(logger)
	
	agent, err := NewGRPCMetricsAgent(cfg, logger, metadataTransport)
	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.NotNil(t, agent.conn)
	assert.NotNil(t, agent.client)
	assert.Equal(t, cfg, agent.cfg)
	assert.Equal(t, logger, agent.logger)
	
	// Закрываем соединение
	err = agent.Close()
	assert.NoError(t, err)
}

func TestGRPCErrorClassifier_Classify(t *testing.T) {
	classifier := &GRPCErrorClassifier{}
	err := errors.New("test error")
	classification := classifier.Classify(err)
	assert.Equal(t, retry.Retriable, classification)
}
