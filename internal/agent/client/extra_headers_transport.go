package agent

import (
	"net"
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
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		et.Logger.Debug("failed to get interface addresses", zap.Error(err))
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if !ipnet.IP.IsLinkLocalUnicast() {
					return ipnet.IP.String()
				}
			}
		}
	}

	et.Logger.Warn("Only link-local IP found, consider checking network setup")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String() // fallback
			}
		}
	}

	return ""
}
