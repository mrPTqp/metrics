package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

// Middleware для проверки IP клиента на основании заголовка X-Real-IP.
// Если IP не входит в маску trustedSubnet, то возвращается 403 Forbidden
func SubnetMiddleware(trustedSubnet *net.IPNet, baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIp := r.Header.Get("X-Real-IP")
			if clientIp == "" {
				baseLogger.Error("request has empty X-Real-IP header")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !trustedSubnet.Contains(net.ParseIP(clientIp)) {
				baseLogger.Error("request has untrusted IP in X-Real-IP header")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
