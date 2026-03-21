package agent

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"net/http"
	"testing"

	"go.uber.org/zap/zaptest"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestEncryptTransport_RoundTrip_EncryptsBody(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	cert := &x509.Certificate{
		PublicKey: &privateKey.PublicKey,
	}

	plainBody := []byte("plain body")

	var capturedBody []byte
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("RoundTripper read body error = %v", err)
		}
		capturedBody = append([]byte(nil), body...)

		// return dummy response
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
		}
		return resp, nil
	})

	et := &EncryptTransport{
		RoundTripper: rt,
		Cert:         cert,
		Logger:       zaptest.NewLogger(t),
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(plainBody))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	resp, err := et.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v, want nil", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	defer resp.Body.Close()

	if bytes.Equal(capturedBody, plainBody) {
		t.Errorf("expected encrypted body to differ from plain body")
	}
	if req.ContentLength != int64(len(capturedBody)) {
		t.Errorf("ContentLength = %d, want %d", req.ContentLength, len(capturedBody))
	}
}

