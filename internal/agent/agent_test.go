package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrPTqp/metrics/internal/models"
)

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
				if ctype := r.Header.Get("Content-Type"); ctype != "application/json" {
					t.Errorf("expected application/json, got %s", ctype)
				}

				body := r.Body
				if r.Header.Get("Content-Encoding") == "gzip" {
					gz, err := gzip.NewReader(r.Body)
					if err != nil {
						t.Errorf("failed to create gzip reader: %v", err)
						return
					}
					defer gz.Close()
					body = gz
				}

				data, err := io.ReadAll(body)
				if err != nil {
					t.Errorf("failed to read body: %v", err)
					return
				}

				var req models.Metrics
				if err := json.Unmarshal(data, &req); err != nil {
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

func ptr[T any](v T) *T { return &v }
