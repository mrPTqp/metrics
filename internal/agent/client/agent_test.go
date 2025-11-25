package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/signer"
)

func TestMetricsAgent_SendMetrics(t *testing.T) {
	tests := []struct {
		name           string
		gauges         map[string]float64
		counters       map[string]int64
		serverStatus   int
		serverResponse string
		wantErr        bool
		expectRequest  bool
		secretKey      string
	}{
		{
			name: "successful send with both gauges and counters",
			gauges: map[string]float64{
				"Alloc":       123.45,
				"RandomValue": 0.987,
			},
			counters: map[string]int64{
				"PollCount": 42,
			},
			serverStatus:   http.StatusOK,
			serverResponse: `{"status":"ok"}`,
			wantErr:        false,
			expectRequest:  true,
			secretKey:      "",
		},
		{
			name:           "empty metrics - no request should be sent",
			gauges:         map[string]float64{},
			counters:       map[string]int64{},
			serverStatus:   http.StatusOK,
			serverResponse: "",
			wantErr:        false,
			expectRequest:  false,
			secretKey:      "",
		},
		{
			name: "only gauges",
			gauges: map[string]float64{
				"HeapAlloc": 678.90,
				"GCSys":     1123.45,
			},
			counters:      map[string]int64{},
			serverStatus:  http.StatusOK,
			wantErr:       false,
			expectRequest: true,
			secretKey:     "",
		},
		{
			name:   "only counters",
			gauges: map[string]float64{},
			counters: map[string]int64{
				"PollCount": 1,
				"Requests":  100,
			},
			serverStatus:  http.StatusOK,
			wantErr:       false,
			expectRequest: true,
			secretKey:     "",
		},
		{
			name: "server returns 500 error",
			gauges: map[string]float64{
				"Alloc": 123.45,
			},
			counters: map[string]int64{
				"PollCount": 1,
			},
			serverStatus:   http.StatusInternalServerError,
			serverResponse: `{"error":"internal server error"}`,
			wantErr:        true,
			expectRequest:  true,
			secretKey:      "",
		},
		{
			name: "server returns 400 bad request",
			gauges: map[string]float64{
				"Alloc": 123.45,
			},
			counters: map[string]int64{
				"PollCount": 1,
			},
			serverStatus:   http.StatusBadRequest,
			serverResponse: `{"error":"bad request"}`,
			wantErr:        true,
			expectRequest:  true,
			secretKey:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestReceived bool
			var requestBody []byte
			var requestHeaders http.Header

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestReceived = true
				requestHeaders = r.Header.Clone()

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}
				requestBody = body

				if r.URL.Path != "/updates" {
					t.Errorf("Expected path /updates, got %s", r.URL.Path)
				}
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					_, _ = w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			var secretKey *string
			if tt.secretKey != "" {
				secretKey = &tt.secretKey
			}

			cfg := &config.Config{
				Address: models.NetAddress{
					Host: "http://localhost",
					Port: 8888,
				},
				SecretKey:      secretKey,
				ReportInterval: 10,
				PoolInterval:   2,
			}

			client := &http.Client{}
			agent := NewMetricsAgent(client, cfg)

			err := agent.sendMetrics(tt.gauges, tt.counters, server.URL[len("http://"):])

			if (err != nil) != tt.wantErr {
				t.Errorf("sendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}

			if requestReceived != tt.expectRequest {
				t.Errorf("Request received = %v, expectRequest %v", requestReceived, tt.expectRequest)
			}

			// Проверка только если запрос был отправлен и не ожидается ошибка
			if requestReceived && tt.expectRequest && !tt.wantErr {
				if contentType := requestHeaders.Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Content-Type header = %s, expected application/json", contentType)
				}
				if contentEncoding := requestHeaders.Get("Content-Encoding"); contentEncoding != "gzip" {
					t.Errorf("Content-Encoding header = %s, expected gzip", contentEncoding)
				}

				if len(requestBody) == 0 {
					t.Fatal("Request body is empty")
				}

				decompressedBody, err := Decompress(requestBody)
				if err != nil {
					t.Fatalf("Failed to decompress request body: %v", err)
				}

				var metrics []models.Metrics
				if err := json.Unmarshal(decompressedBody, &metrics); err != nil {
					t.Fatalf("Invalid JSON in request: %v", err)
				}

				expectedCount := len(tt.gauges) + len(tt.counters)
				if len(metrics) != expectedCount {
					t.Errorf("Number of metrics = %d, expected %d", len(metrics), expectedCount)
				}

				gaugeCount := 0
				counterCount := 0
				for _, metric := range metrics {
					switch metric.MType {
					case "gauge":
						gaugeCount++
						if metric.Value == nil {
							t.Error("Gauge metric has nil Value")
						}
					case "counter":
						counterCount++
						if metric.Delta == nil {
							t.Error("Counter metric has nil Delta")
						}
					default:
						t.Errorf("Unknown metric type: %s", metric.MType)
					}
				}

				if gaugeCount != len(tt.gauges) {
					t.Errorf("Gauge count = %d, expected %d", gaugeCount, len(tt.gauges))
				}
				if counterCount != len(tt.counters) {
					t.Errorf("Counter count = %d, expected %d", counterCount, len(tt.counters))
				}
			}
		})
	}
}

func TestMetricsAgent_SendMetrics_RequestStructure(t *testing.T) {
	tests := []struct {
		name     string
		gauges   map[string]float64
		counters map[string]int64
	}{
		{
			name: "single gauge and counter",
			gauges: map[string]float64{
				"TestGauge": 123.456,
			},
			counters: map[string]int64{
				"TestCounter": 789,
			},
		},
		{
			name: "multiple metrics",
			gauges: map[string]float64{
				"Gauge1": 1.1,
				"Gauge2": 2.2,
				"Gauge3": 3.3,
			},
			counters: map[string]int64{
				"Counter1": 10,
				"Counter2": 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedMetrics []models.Metrics

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				decompressedBody, _ := Decompress(body)
				_ = json.Unmarshal(decompressedBody, &receivedMetrics)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			cfg := &config.Config{
				Address: models.NetAddress{
					Host: "http://localhost",
					Port: 8888,
				},
			}

			client := &http.Client{}
			agent := NewMetricsAgent(client, cfg)

			err := agent.sendMetrics(tt.gauges, tt.counters, server.URL[len("http://"):])
			if err != nil {
				t.Fatalf("sendMetrics() unexpected error: %v", err)
			}

			if len(receivedMetrics) != len(tt.gauges)+len(tt.counters) {
				t.Errorf("Expected %d metrics, got %d", len(tt.gauges)+len(tt.counters), len(receivedMetrics))
			}

			for name, expectedValue := range tt.gauges {
				found := false
				for _, metric := range receivedMetrics {
					if metric.ID == name && metric.MType == "gauge" && metric.Value != nil && *metric.Value == expectedValue {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Gauge metric %s not found or has wrong value", name)
				}
			}

			for name, expectedDelta := range tt.counters {
				found := false
				for _, metric := range receivedMetrics {
					if metric.ID == name && metric.MType == "counter" && metric.Delta != nil && *metric.Delta == expectedDelta {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Counter metric %s not found or has wrong delta", name)
				}
			}
		})
	}
}

func TestMetricsAgent_SendMetrics_WithSecretKey(t *testing.T) {
	secretKey := "mysecret"
	cfg := &config.Config{
		Address: models.NetAddress{
			Host: "http://localhost",
			Port: 8888,
		},
		SecretKey: &secretKey,
	}

	client := &http.Client{}
	agent := NewMetricsAgent(client, cfg)

	var receivedHash string
	var requestBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHash = r.Header.Get("HashSHA256")
		body, _ := io.ReadAll(r.Body)
		requestBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	gauges := map[string]float64{"TestGauge": 123.45}
	counters := map[string]int64{"TestCounter": 1}

	err := agent.sendMetrics(gauges, counters, server.URL[len("http://"):])
	if err != nil {
		t.Fatalf("sendMetrics failed: %v", err)
	}

	if receivedHash == "" {
		t.Fatal("Expected HashSHA256 header, but it's missing")
	}

	expectedHash, err := sign.Sign(requestBody, &secretKey)
	if err != nil {
		t.Fatalf("Failed to compute expected hash: %v", err)
	}

	if receivedHash != *expectedHash {
		t.Errorf("HashSHA256 mismatch:\nexpected: %s\ngot:      %s", *expectedHash, receivedHash)
	}
}
