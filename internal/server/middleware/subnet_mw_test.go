package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestSubnetMiddleware(t *testing.T) {
	logger := zaptest.NewLogger(t)

	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	middleware := SubnetMiddleware(subnet, logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	tests := []struct {
		name           string
		clientIP       string
		expectedStatus int
	}{
		{
			name:           "Trusted IP in subnet",
			clientIP:       "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Trusted IP at subnet boundary",
			clientIP:       "192.168.1.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Trusted IP at subnet boundary",
			clientIP:       "192.168.1.255",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Untrusted IP outside subnet",
			clientIP:       "192.168.2.100",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Untrusted IP outside subnet",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Empty X-Real-IP header",
			clientIP:       "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
