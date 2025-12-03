// internal/server/middleware/sign_mw.go
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/signer"
	"go.uber.org/zap"
)

func writeJSONError(w http.ResponseWriter, message string, statusCode int, logger *zap.SugaredLogger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func SignMiddleware(h http.HandlerFunc, key string, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorw("failed to read request body", "error", err)
			writeJSONError(w, "invalid request body", http.StatusBadRequest, logger)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(bodyContent))

		signHeader := r.Header.Get("HashSHA256")
		if signHeader != "" {
			if !signer.Verify(bodyContent, &signHeader, &key, logger) {
				logger.Warn("invalid signature in request")
				writeJSONError(w, "invalid signature", http.StatusBadRequest, logger)
				return
			}
		}

		ww := &signingResponseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		h.ServeHTTP(ww, r)

		responseBody := ww.body.Bytes()
		sign, err := signer.Sign(responseBody, &key)
		if err != nil {
			logger.Errorw("failed to sign response", "error", err)
			return
		}

		ww.ResponseWriter.Header().Set("HashSHA256", *sign)
	}
}

type signingResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (rw *signingResponseWriter) Write(b []byte) (int, error) {
	// Сохраняем тело для подписи
	_, err := rw.body.Write(b)
	if err != nil {
		return 0, err
	}
	// Отправляем клиенту
	return rw.ResponseWriter.Write(b)
}

func (rw *signingResponseWriter) WriteHeader(code int) {
	rw.ResponseWriter.WriteHeader(code)
}
