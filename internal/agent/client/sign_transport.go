package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/signer"
	"go.uber.org/zap"
)

// Transport для подписи запросов и проверки подписи ответов
type SigningTransport struct {
	RoundTripper http.RoundTripper
	SecretKey    string
	Logger       *zap.Logger
}

// Подписываем запрос и проверяем подпись ответа
func (st *SigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело
	}

	if len(bodyBytes) > 0 {
		signature, err2 := signer.Sign(bodyBytes, &st.SecretKey)
		if err2 != nil {
			return nil, err2
		}
		req.Header.Set("HashSHA256", *signature)
	}

	resp, err := st.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	signature := resp.Header.Get("HashSHA256")
	if signature == "" {
		return resp, nil
	}
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		st.Logger.Warn("failed to read response body", zap.Error(err))
		return resp, nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело

	if !signer.Verify(bodyBytes, &signature, &st.SecretKey, st.Logger) {
		return nil, fmt.Errorf("response signature verification failed")
	}

	return resp, err
}
