package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestLoggingMiddleware_SuccessAndErrorStatuses(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		status         int
		wantStatus     int
		withReqBody    bool
		withPanic      bool
		expectHTTPCode int
	}{
		{
			name:           "successful request 200",
			status:         http.StatusOK,
			wantStatus:     http.StatusOK,
			withReqBody:    true,
			expectHTTPCode: http.StatusOK,
		},
		{
			name:           "client error 400",
			status:         http.StatusBadRequest,
			wantStatus:     http.StatusBadRequest,
			withReqBody:    true,
			expectHTTPCode: http.StatusBadRequest,
		},
		{
			name:           "server error 500",
			status:         http.StatusInternalServerError,
			wantStatus:     http.StatusInternalServerError,
			withReqBody:    false,
			expectHTTPCode: http.StatusInternalServerError,
		},
		{
			name:           "panic recovered",
			status:         http.StatusOK,
			withReqBody:    true,
			withPanic:      true,
			expectHTTPCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.withPanic {
					panic("test panic")
				}
				if tt.status != 0 {
					w.WriteHeader(tt.status)
				}
				_, _ = w.Write([]byte("response body"))
			})

			logMw := LoggingMiddleware(logger)
			h := logMw(baseHandler)

			var reqBody *bytes.Buffer
			if tt.withReqBody {
				reqBody = bytes.NewBufferString("request body")
			} else {
				reqBody = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/", reqBody)
			req.RemoteAddr = "127.0.0.1:1234"
			req.Header.Set("User-Agent", "test-agent")

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.expectHTTPCode {
				t.Errorf("response code = %d, want %d", rr.Code, tt.expectHTTPCode)
			}
		})
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "from X-Forwarded-For first IP",
			remoteAddr: "10.0.0.1:1234",
			xForwardedFor: "1.2.3.4, 5.6.7.8",
			expectedIP: "1.2.3.4",
		},
		{
			name:       "from X-Real-IP",
			remoteAddr: "10.0.0.1:1234",
			xRealIP:    "9.8.7.6",
			expectedIP: "9.8.7.6",
		},
		{
			name:       "from RemoteAddr",
			remoteAddr: "192.168.0.1:8080",
			expectedIP: "192.168.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

