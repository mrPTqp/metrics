package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тест sendMetric с параметризацией различных HTTP-статусов
func TestSendMetric(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expectErr  bool
	}{
		{"Success", http.StatusOK, false},
		{"ServerError", http.StatusInternalServerError, true},
		{"NotFound", http.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			path := "/update/gauge/testmetric/123.45"
			url := server.URL + path
			err := sendMetric(http.Client{}, url)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
