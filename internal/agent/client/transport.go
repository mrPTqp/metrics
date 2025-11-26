package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/mrPTqp/metrics/internal/signer"
	"go.uber.org/zap"
)

type SigningTransport struct {
	RoundTripper http.RoundTripper
	SecretKey    string
	Logger       *zap.SugaredLogger
}

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
		signature, err := signer.Sign(bodyBytes, &st.SecretKey)
		if err != nil {
			return nil, err
		}
		req.Header.Set("HashSHA256", *signature)
	}

	resp, err := st.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело ответа

		if err != nil {
			st.Logger.Warnw("failed to read response body", "error", err)
			return resp, nil
		}

		if len(bodyBytes) > 0 {
			signature := resp.Header.Get("HashSHA256")
			if signature == "" {
				st.Logger.Warn("missing HashSHA256 header in response")
				return resp, nil
			}
			if !signer.Verify(bodyBytes, &signature, &st.SecretKey, st.Logger) {
				return nil, fmt.Errorf("response signature verification failed")
			}
		}
	}

	return resp, err
}
