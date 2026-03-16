package handler

import (
	"context"
	"net"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// SubnetInterceptor проверяет IP-адрес клиента на принадлежность доверенной подсети
func SubnetInterceptor(trustedSubnet *net.IPNet, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet == nil {
			// Если подсеть не настроена, пропускаем проверку
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("no metadata in context")
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		realIPs := md.Get("x-real-ip")
		if len(realIPs) == 0 {
			logger.Warn("x-real-ip header not found")
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		clientIPStr := strings.TrimSpace(realIPs[0])
		clientIP := net.ParseIP(clientIPStr)
		if clientIP == nil {
			logger.Warn("invalid x-real-ip", zap.String("ip", clientIPStr))
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		if !trustedSubnet.Contains(clientIP) {
			logger.Warn("client IP not in trusted subnet", zap.String("ip", clientIPStr), zap.String("subnet", trustedSubnet.String()))
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		logger.Debug("client IP verified", zap.String("ip", clientIPStr))
		return handler(ctx, req)
	}
}