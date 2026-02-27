package middleware

import (
	"bytes"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/mrPTqp/metrics/internal/contextkey"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status   int
	size     int
	hijacked bool
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.size += n
	return n, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	if lrw.status == 0 {
		lrw.status = statusCode
	}
	lrw.ResponseWriter.WriteHeader(statusCode)
}

// Middleware для логирования HTTP-запросов и ответов. Запрос и ответ логируются одним логом с контекстом
func LoggingMiddleware(baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := uuid.New().String()
			ip := getIP(r)
			ctx := contextkey.WithClientIP(r.Context(), ip)

			start := time.Now()

			log := baseLogger.With(
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("rid", rid),
			)

			ctx = contextkey.WithLogger(ctx, log)
			r = r.WithContext(ctx)

			bodyBuf := &bytes.Buffer{}
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
				hijacked:       false,
			}

			defer func() {
				log := contextkey.LoggerFromContext(ctx)
				duration := time.Since(start)

				status := lrw.status
				if status == 0 {
					status = 200
				}

				log = log.With(
					zap.Int("status", status),
					zap.Int("size", lrw.size),
					zap.Duration("duration", duration),
				)

				var reqBody []byte
				if capturedBody := contextkey.RequestBodyFromContext(r.Context()); capturedBody != nil {
					reqBody = capturedBody
					log = log.With(zap.ByteString("request_body", reqBody))
				} else {
					log = log.With(zap.String("request_body", "[not captured]"))
				}

				if p := recover(); p != nil {
					log.Error("panic recovered",
						zap.Any("panic", p),
						zap.ByteString("request_body", reqBody),
						zap.ByteString("response_body", bodyBuf.Bytes()),
					)
					http.Error(lrw, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				if status >= 500 {
					log.Error("server error",
						zap.ByteString("request_body", reqBody),
						zap.ByteString("response_body", bodyBuf.Bytes()),
					)
					return
				}

				if status >= 400 {
					log.Warn("client error",
						zap.ByteString("request_body", reqBody),
						zap.ByteString("response_body", bodyBuf.Bytes()),
					)
					return
				}

				log.Info("successful request")
			}()

			next.ServeHTTP(lrw, r)
		})
	}
}

func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
