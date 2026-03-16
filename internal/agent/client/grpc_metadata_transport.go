package agent

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GRPCMetadataTransport добавляет метаданные к gRPC запросам
type GRPCMetadataTransport struct {
	logger *zap.Logger
}

// NewGRPCMetadataTransport создает новый транспорт для gRPC метаданных
func NewGRPCMetadataTransport(logger *zap.Logger) *GRPCMetadataTransport {
	return &GRPCMetadataTransport{
		logger: logger,
	}
}

// UnaryInterceptor добавляет метаданные к unary gRPC запросам
func (gmt *GRPCMetadataTransport) UnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ip := GetClientIP(gmt.logger)
		if ip != "" {
			gmt.logger.Debug("Adding client IP to gRPC metadata", zap.String("ip", ip))
			
			md := metadata.Pairs("x-real-ip", ip)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

