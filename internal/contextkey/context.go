package contextkey

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

var ClientIPKey = &struct{ name string }{"client-ip"}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPKey, ip)
}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return log
	}
	return zap.L() // fallback
}

func GetClientIP(ctx context.Context) string {
	ip := "unknown"
	if clientIP, ok := ctx.Value(ClientIPKey).(string); ok {
		ip = clientIP
	}
	return ip
}
