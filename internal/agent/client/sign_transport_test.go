package agent

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/signer"
)

func TestSigningTransport_RoundTrip_SignsRequestAndVerifiesResponse(t *testing.T) {
	secret := "secret-key"
	logger := zaptest.NewLogger(t)

	// RoundTripper проверяет подпись запроса и выставляет корректную подпись ответа
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body in RT: %v", err)
		}

		// проверяем, что запрос подписан
		sigHeader := req.Header.Get("HashSHA256")
		if sigHeader == "" {
			t.Fatalf("request HashSHA256 header is empty, want non-empty")
		}

		if !signer.Verify(body, &sigHeader, &secret, logger) {
			t.Fatalf("request signature verification failed in RT")
		}

		// ответ с подписью
		respBody := []byte("response body")
		respSig, err := signer.Sign(respBody, &secret)
		if err != nil {
			t.Fatalf("failed to sign response body in RT: %v", err)
		}

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(respBody)),
		}
		resp.Header.Set("HashSHA256", *respSig)
		return resp, nil
	})

	st := &SigningTransport{
		RoundTripper: rt,
		SecretKey:    secret,
		Logger:       logger,
	}

	body := []byte("request body")
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(body))

	resp, err := st.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v, want nil", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	defer resp.Body.Close()

	// тело ответа должно быть доступно после проверки подписи
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body after RoundTrip: %v", err)
	}
	if string(respBody) != "response body" {
		t.Errorf("response body = %q, want %q", string(respBody), "response body")
	}
}

func TestSigningTransport_RoundTrip_ResponseWithoutSignature(t *testing.T) {
	secret := "secret-key"
	logger := zaptest.NewLogger(t)

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader([]byte("no sig"))),
		}, nil
	})

	st := &SigningTransport{
		RoundTripper: rt,
		SecretKey:    secret,
		Logger:       logger,
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := st.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v, want nil", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	defer resp.Body.Close()
}

