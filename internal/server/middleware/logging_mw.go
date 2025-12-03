package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggingMiddleware(h http.HandlerFunc, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var body []byte
        if r.Body != nil {
            body, _ = io.ReadAll(r.Body)
            r.Body = io.NopCloser(bytes.NewReader(body)) // Восстанавливаем
        }

		responseData := &responseData{
			status: -1,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		 logger.Infow("incoming request",
            "method", r.Method,
            "uri", r.RequestURI,
            "headers", r.Header,
            "body", string(body),
        )

		defer func() {
			if p := recover(); p != nil {
				var msg string
				switch v := p.(type) {
				case error:
					msg = v.Error()
				case string:
					msg = v
				default:
					msg = fmt.Sprintf("%v", v)
				}

				custom := fmt.Errorf("panic: %v", p)

				logger.Errorf("PANIC RECOVERED: status=%d, message=%s, error=%v", http.StatusInternalServerError, msg, custom)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			duration := time.Since(start)

			logger.Infoln(
                "handled request",
                "uri", r.RequestURI,
                "method", r.Method,
                "duration", duration,
                "status", responseData.status,
                "size", responseData.size,
            )
		}()

		h.ServeHTTP(&lw, r)
	}
}
