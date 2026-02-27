package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/contextkey"
)

// Middleware для извлечения дешифрованного и распакованного тела запроса для последующего логирования
func LoggingRequestBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := contextkey.LoggerFromContext(r.Context())

		if r.Body == nil || r.ContentLength <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONError(w, "failed to read request body", http.StatusInternalServerError, log)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		ctx := contextkey.WithRequestBody(r.Context(), body)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
