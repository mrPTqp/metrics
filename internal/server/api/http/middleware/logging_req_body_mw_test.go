package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
)

func TestLoggingRequestBodyMiddleware(t *testing.T) {
	logger := zaptest.NewLogger(t)

	var capturedBody []byte
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// тело должно быть доступно и в r.Body, и в контексте
		bodyFromCtx := contextkey.RequestBodyFromContext(r.Context())
		if bodyFromCtx != nil {
			capturedBody = append([]byte(nil), bodyFromCtx...)
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := LoggingRequestBodyMiddleware(baseHandler)

	t.Run("body captured and restored", func(t *testing.T) {
		originalBody := []byte("test request body")

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(originalBody))
		req.ContentLength = int64(len(originalBody))
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if !bytes.Equal(capturedBody, originalBody) {
			t.Errorf("captured body = %q, want %q", string(capturedBody), string(originalBody))
		}
	})

	t.Run("no body - middleware passes through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))
		rr := httptest.NewRecorder()

		mw.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})
}

