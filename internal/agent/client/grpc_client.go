package agent

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/agent/repository"
	"github.com/mrPTqp/metrics/internal/proto"
	"github.com/mrPTqp/metrics/internal/retry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// GRPCMetricsAgent отправляет метрики на сервер по gRPC
type GRPCMetricsAgent struct {
	conn   *grpc.ClientConn
	client proto.MetricsClient
	cfg    *config.Config
	logger *zap.Logger
}

// NewGRPCMetricsAgent создает новый экземпляр GRPCMetricsAgent
func NewGRPCMetricsAgent(cfg *config.Config, logger *zap.Logger, metadataTransport *GRPCMetadataTransport) (*GRPCMetricsAgent, error) {
	conn, err := grpc.NewClient(
		cfg.GRPCServerAddress.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(metadataTransport.UnaryInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewMetricsClient(conn)

	return &GRPCMetricsAgent{
		conn:   conn,
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

// SendMetrics отправляет собранные метрики на сервер по gRPC
func (ga *GRPCMetricsAgent) SendMetrics(ctx context.Context, repo repository.MetricRepository) {
	gauges, counters := repo.GetAllMetrics()
	additionalGauges := repo.GetAdditionalGaugeMetrics()

	if len(gauges) == 0 && len(counters) == 0 && len(additionalGauges) == 0 {
		ga.logger.Info("gauges and counters are empty")
		return
	}

	metrics := make([]*proto.Metric, 0, len(gauges)+len(additionalGauges)+len(counters))

	for name, value := range gauges {
		metrics = append(metrics, &proto.Metric{
			Id:    name,
			Type:  proto.Metric_GAUGE,
			Value: &value,
		})
	}

	for name, value := range additionalGauges {
		metrics = append(metrics, &proto.Metric{
			Id:    name,
			Type:  proto.Metric_GAUGE,
			Value: &value,
		})
	}

	for name, delta := range counters {
		metrics = append(metrics, &proto.Metric{
			Id:    name,
			Type:  proto.Metric_COUNTER,
			Delta: &delta,
		})
	}

	req := &proto.UpdateMetricsRequest{
		Metrics: metrics,
	}


	err := retry.DoWithRetry(
		ctx,
		&GRPCErrorClassifier{},
		func() error {
			_, err := ga.client.UpdateMetrics(ctx, req)
			return err
		},
		3,
		1*time.Second,
	)

	if err != nil {
		ga.logger.Error("failed to send metrics via gRPC after retries", zap.Error(err))
		return
	}

	ga.logger.Info("metrics successfully sent via gRPC")
}

// Close закрывает соединение
func (ga *GRPCMetricsAgent) Close() error {
	return ga.conn.Close()
}


// GRPCErrorClassifier классифицирует ошибки gRPC для retry
type GRPCErrorClassifier struct{}

func (ec *GRPCErrorClassifier) Classify(err error) retry.ErrorClassification {
	if err == nil {
		return retry.NonRetriable
	}

	if err.Error() == "context canceled" {
		return retry.NonRetriable
	}

	grpcStatus, ok := status.FromError(err)
	if !ok {
		return retry.Retriable
	}

	switch grpcStatus.Code() {
	case codes.InvalidArgument, codes.Unauthenticated, codes.PermissionDenied:
		return retry.NonRetriable
	case codes.Unavailable, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.ResourceExhausted:
		return retry.Retriable
	case codes.Canceled:
		return retry.NonRetriable
	default:
		return retry.Retriable
	}
}
