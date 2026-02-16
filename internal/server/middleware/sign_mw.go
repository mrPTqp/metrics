package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/signer"
	"go.uber.org/zap"
)

func writeJSONError(w http.ResponseWriter, message string, statusCode int, logger *zap.Logger) {
	logger.Warn("Sending JSON error response",
		zap.Int("status", statusCode),
		zap.String("error", message))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Middleware для проверки подписи запроса и генерации подписи ответа.
// Key - секретный ключ для подписи (одинаковый для запроса и ответа).
// Подпись генерируется в заголовке "HashSHA256"
func SignMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var bodyContent []byte
			log := contextkey.LoggerFromContext(r.Context())
			var err error

			bodyContent, err = io.ReadAll(r.Body)
			if err != nil {
				log.Error("failed to read request body", zap.Error(err))
				writeJSONError(w, "invalid request body", http.StatusBadRequest, log)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyContent))

			signHeader := r.Header.Get("HashSHA256")
			if signHeader != "" {
				if !signer.Verify(bodyContent, &signHeader, &key, log) {
					log.Warn("request signature verification failed",
						zap.String("url", r.URL.String()),
						zap.String("method", r.Method))
					writeJSONError(w, "invalid signature", http.StatusBadRequest, log)
					return
				}
				log.Debug("request signature verified successfully",
					zap.String("method", r.Method),
					zap.String("url", r.URL.String()))
			}

			ww := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
			next.ServeHTTP(ww, r)

			responseBody := ww.body.Bytes()
			sign, err := signer.Sign(responseBody, &key)
			if err != nil {
				log.Error("failed to sign response", zap.Error(err))
				writeJSONError(w, "failed to sign response", http.StatusInternalServerError, log)
				return
			}

			log.Debug("response signed successfully",
				zap.String("hash", *sign),
				zap.Int("body_size", len(responseBody)))

			ww.Header().Set("HashSHA256", *sign)

			for k, v := range ww.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(ww.code)
			_, _ = w.Write(ww.body.Bytes())
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}
