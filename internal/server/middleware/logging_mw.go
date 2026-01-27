package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/mrPTqp/metrics/internal/contextkey"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	bodyBuf  *bytes.Buffer
	status   int
	size     int
	hijacked bool
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.size += n
	lrw.bodyBuf.Write(b)
	return n, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	if lrw.status == 0 {
		lrw.status = statusCode
	}
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func readBody(r io.ReadCloser) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	r.Close()
	return buf.Bytes(), nil
}

func LoggingMiddleware(baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := uuid.New().String()
			start := time.Now()

			log := baseLogger.With(
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("rid", rid),
			)

			ctx := contextkey.WithLogger(r.Context(), log)
			r = r.WithContext(ctx)

			var reqBody []byte
			if r.Body != nil && r.ContentLength > 0 {
				var err error
				reqBody, err = readBody(r.Body)
				if err != nil {
					log.Error("Failed to read request body", zap.Error(err))
				} else {
					r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
				}
			}

			bodyBuf := &bytes.Buffer{}
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				bodyBuf:        bodyBuf,
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

				if p := recover(); p != nil {
					log.Error("Panic recovered",
						zap.Any("panic", p),
						zap.ByteString("request_body", reqBody),
						zap.ByteString("response_body", bodyBuf.Bytes()),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				
				if status >= 500 {
					log.Error("Server error",
						zap.ByteString("request_body", reqBody),
						zap.ByteString("response_body", bodyBuf.Bytes()),
					)
					return
				}

				if status >= 400 {
					log.Warn("Client error",
						zap.ByteString("request_body", reqBody),
						zap.ByteString("response_body", bodyBuf.Bytes()),
					)
					return
				}

				log.Info("Successful request")
			}()

			next.ServeHTTP(lrw, r)
		})
	}
}
