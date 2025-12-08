// internal/server/middleware/sign_mw.go
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

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
		logger.Infow("incoming request body", "body", string(bodyContent), "method", r.Method, "url", r.URL.Path)

		r.Body = io.NopCloser(bytes.NewReader(bodyContent))

		var signHeader string
		for name, values := range r.Header {
			if strings.EqualFold(name, "HashSHA256") {
				if len(values) > 0 {
					signHeader = values[0]
				}
				break
			}
		}

		logger.Infow("received headers", "headers", r.Header)
		logger.Infow("looking for HashSHA256", "value", signHeader)

		if signHeader != "" {
			if !signer.Verify(bodyContent, signHeader, key, logger) {
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
		logger.Infow("outgoing response body", "body", string(responseBody))

		sign, err := signer.Sign(responseBody, key)
		if err != nil {
			logger.Errorw("failed to sign response", "error", err)
			return
		}

		ww.ResponseWriter.Header().Set("HashSHA256", sign)
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
