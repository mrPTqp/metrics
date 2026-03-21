package middleware

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/crypto"
	"go.uber.org/zap"
)

// Middleware для дешифровки тела ответа приватным ключом
func DecryptMiddleware(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}
			defer r.Body.Close()

			log := contextkey.LoggerFromContext(r.Context())

			payload, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			decrypted, err := crypto.Decrypt(privateKey, payload)
			if err != nil {
				log.Error("failed to decrypt request body", zap.Error(err))
				if errors.Is(err, crypto.ErrAESCipherCreate) || errors.Is(err, crypto.ErrGCMCreate) {
					writeJSONError(w, "Internal Server Error", http.StatusInternalServerError, log)
				} else {
					writeJSONError(w, "invalid request body", http.StatusBadRequest, log)
				}
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))

			next.ServeHTTP(w, r)
		})
	}
}
