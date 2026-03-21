package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/crypto"
)

func TestDecryptMiddleware_SuccessAndErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	cert := &x509.Certificate{
		PublicKey: &privateKey.PublicKey,
	}

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// simply echo decrypted body to confirm it was replaced
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	mw := DecryptMiddleware(privateKey)

	t.Run("GET request bypasses decryption", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))
		rr := httptest.NewRecorder()

		mw(baseHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("successful decryption of POST body", func(t *testing.T) {
		plain := []byte(`{"ok":true}`)
		enc, err := crypto.Encrypt(cert, plain)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(enc))
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))
		rr := httptest.NewRecorder()

		mw(baseHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if !bytes.Equal(rr.Body.Bytes(), plain) {
			t.Errorf("decrypted body = %q, want %q", rr.Body.String(), string(plain))
		}
	})

	t.Run("invalid encrypted payload returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("invalid")))
		req = req.WithContext(contextkey.WithLogger(req.Context(), logger))
		rr := httptest.NewRecorder()

		mw(baseHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest && rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 400 or 500", rr.Code)
		}
	})
}

