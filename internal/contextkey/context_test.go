package contextkey

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestLoggerContext(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name         string
		ctx          context.Context
		withLogger   bool
		expectLogger bool
	}{
		{
			name:         "with logger in context",
			ctx:          WithLogger(context.Background(), logger),
			withLogger:   true,
			expectLogger: true,
		},
		{
			name:         "fallback logger when not set",
			ctx:          context.Background(),
			withLogger:   false,
			expectLogger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoggerFromContext(tt.ctx)
			if tt.expectLogger && got == nil {
				t.Fatal("LoggerFromContext() returned nil")
			}
			if tt.withLogger && got.Core() != logger.Core() {
				t.Errorf("LoggerFromContext() returned different logger core")
			}
		})
	}
}

func TestClientIPContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantIP string
	}{
		{
			name:   "ip set in context",
			ctx:    WithClientIP(context.Background(), "127.0.0.1"),
			wantIP: "127.0.0.1",
		},
		{
			name:   "ip not set",
			ctx:    context.Background(),
			wantIP: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ip := GetClientIP(tt.ctx); ip != tt.wantIP {
				t.Errorf("GetClientIP() = %q, want %q", ip, tt.wantIP)
			}
		})
	}
}

func TestRequestBodyContext(t *testing.T) {
	body := []byte("test body")

	tests := []struct {
		name      string
		ctx       context.Context
		wantBody  []byte
		wantNil   bool
	}{
		{
			name:     "body set in context",
			ctx:      WithRequestBody(context.Background(), body),
			wantBody: body,
			wantNil:  false,
		},
		{
			name:     "body not set",
			ctx:      context.Background(),
			wantBody: nil,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequestBodyFromContext(tt.ctx)
			if tt.wantNil {
				if got != nil {
					t.Errorf("RequestBodyFromContext() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("RequestBodyFromContext() returned nil")
			}
			if string(got) != string(tt.wantBody) {
				t.Errorf("RequestBodyFromContext() = %q, want %q", string(got), string(tt.wantBody))
			}
		})
	}
}


