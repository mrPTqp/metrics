package agent

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/crypto"
	"go.uber.org/zap"
)

type EncryptTransport struct {
	RoundTripper http.RoundTripper
	Cert         *x509.Certificate
	Logger       *zap.Logger
}

func (et *EncryptTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}

		encryptedBody, err := crypto.Encrypt(et.Cert, body)
		if err != nil {
			et.Logger.Error("failed to encrypt body", zap.Error(err))
			return nil, err
		}

		req.Body = io.NopCloser(bytes.NewBuffer(encryptedBody))
		req.ContentLength = int64(len(encryptedBody))
	}

	resp, err := et.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	return resp, err
}
