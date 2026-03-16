package agent

import (
	"net/http"

	"go.uber.org/zap"
)

type ExtraHeadersTransport struct {
	RoundTripper http.RoundTripper
	Logger       *zap.Logger
}

func (et *ExtraHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ip := et.getClientIP()
	if ip != "" {
		et.Logger.Info("Client IP", zap.String("ip", ip))
		req.Header.Set("X-Real-IP", ip)
	}

	return et.RoundTripper.RoundTrip(req)
}

func (et *ExtraHeadersTransport) getClientIP() string {
	return GetClientIP(et.Logger)
}
