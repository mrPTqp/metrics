package agent

import (
	"net"

	"go.uber.org/zap"
)

// GetClientIP получает локальный IP-адрес
func GetClientIP(logger *zap.Logger) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Debug("failed to get interface addresses", zap.Error(err))
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

	logger.Warn("Only link-local IP found, consider checking network setup")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String() // fallback
			}
		}
	}

	return ""
}