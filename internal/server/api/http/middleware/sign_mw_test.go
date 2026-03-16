package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/signer"
)

func TestSignMiddleware_RequestVerificationAndResponseSigning(t *testing.T) {
	secret := "test-secret"
	logger := zaptest.NewLogger(t)

	// Handler just echoes body back
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	mw := SignMiddleware(secret)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// inject logger into context so SignMiddleware can use it
		ctx := contextkey.WithLogger(r.Context(), logger)
		baseHandler.ServeHTTP(w, r.WithContext(ctx))
	}))

	body := []byte(`{"foo":"bar"}`)

	t.Run("valid request signature and response signing", func(t *testing.T) {
		// сначала подписываем тело так же, как это сделает клиент
		sig, err := signer.Sign(body, &secret)
		if err != nil {
			t.Fatalf("failed to sign body for request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
		req.Header.Set("HashSHA256", *sig)
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Header().Get("HashSHA256"); got == "" {
			t.Errorf("response HashSHA256 header is empty, want non-empty")
		}
		if !bytes.Equal(rr.Body.Bytes(), body) {
			t.Errorf("response body = %q, want %q", rr.Body.String(), string(body))
		}
	})

	t.Run("missing or invalid request signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		// без заголовка HashSHA256 middleware не должен отвергать запрос
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})
}

