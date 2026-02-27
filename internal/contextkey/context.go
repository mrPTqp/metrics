package contextkey

import (
	"context"
	"go.uber.org/zap"
)

type loggerKey struct{}
type clientIPKey struct{}
type requestBodyKey struct{}

// Установка логгера в контекст
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Установка IP клиента в контекст
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// WithRequestBody добавляет тело запроса в контекст
func WithRequestBody(ctx context.Context, body []byte) context.Context {
	return context.WithValue(ctx, requestBodyKey{}, body)
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
	if clientIP, ok := ctx.Value(clientIPKey{}).(string); ok {
		ip = clientIP
	}
	return ip
}

// RequestBodyFromContext извлекает тело запроса из контекста
func RequestBodyFromContext(ctx context.Context) []byte {
	if body, ok := ctx.Value(requestBodyKey{}).([]byte); ok {
		return body
	}
	return nil
}