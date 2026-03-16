package handler

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSubnetInterceptor(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name           string
		trustedSubnet  string
		clientIP       string
		metadata       map[string]string
		expectedCode   codes.Code
		expectedLogMsg string
	}{
		{
			name:          "no trusted subnet configured",
			trustedSubnet: "",
			clientIP:      "192.168.1.100",
			metadata: map[string]string{
				"x-real-ip": "192.168.1.100",
			},
			expectedCode: codes.OK,
		},
		{
			name:          "valid IP in trusted subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "192.168.1.100",
			metadata: map[string]string{
				"x-real-ip": "192.168.1.100",
			},
			expectedCode: codes.OK,
		},
		{
			name:          "IP not in trusted subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "10.0.0.1",
			metadata: map[string]string{
				"x-real-ip": "10.0.0.1",
			},
			expectedCode:   codes.PermissionDenied,
			expectedLogMsg: "client IP not in trusted subnet",
		},
		{
			name:          "no x-real-ip header",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "192.168.1.100",
			metadata:      map[string]string{},
			expectedCode:   codes.PermissionDenied,
			expectedLogMsg: "x-real-ip header not found",
		},
		{
			name:          "invalid IP address",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "invalid-ip",
			metadata: map[string]string{
				"x-real-ip": "invalid-ip",
			},
			expectedCode:   codes.PermissionDenied,
			expectedLogMsg: "invalid x-real-ip",
		},
		{
			name:          "empty x-real-ip header",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "",
			metadata: map[string]string{
				"x-real-ip": "",
			},
			expectedCode:   codes.PermissionDenied,
			expectedLogMsg: "invalid x-real-ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trustedSubnet *net.IPNet
			if tt.trustedSubnet != "" {
				_, trustedSubnet, err = net.ParseCIDR(tt.trustedSubnet)
				require.NoError(t, err)
			}

			interceptor := SubnetInterceptor(trustedSubnet, logger)

			// Создаем контекст с метаданными
			md := metadata.New(tt.metadata)
			ctx := metadata.NewIncomingContext(context.Background(), md)

			// Создаем mock handler
			handler := func(ctx context.Context, req any) (interface{}, error) {
				return "success", nil
			}

			// Вызываем interceptor
			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			// Проверяем результат
			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, "success", resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}