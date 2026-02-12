package contextkey

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

var ClientIPKey = &struct{ name string }{"client-ip"}

// Установка логгера контекст
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Установка IP клиента в контекст
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPKey, ip)
}

// Получение логгера из контекста
func LoggerFromContext(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return log
	}
	return zap.L() // fallback
}

// Получение IP клиента из контекста
func GetClientIP(ctx context.Context) string {
	ip := "unknown"
	if clientIP, ok := ctx.Value(ClientIPKey).(string); ok {
		ip = clientIP
	}
	return ip
}
