package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrPTqp/metrics/internal/models"
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
				// Проверяем, что Content-Type правильный
				if ctype := r.Header.Get("Content-Type"); ctype != "application/json" {
					t.Errorf("expected application/json, got %s", ctype)
				}

				// Читаем тело
				body, _ := io.ReadAll(r.Body)
				var req models.Metrics
				if err := json.Unmarshal(body, &req); err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
				}

				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			req := models.Metrics{
				ID:    "test",
				MType: "gauge",
				Value: ptr(123.45),
			}
			err := sendMetric(http.Client{}, server.URL+"/update", req)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

// Вспомогательная функция
func ptr[T any](v T) *T { return &v }
