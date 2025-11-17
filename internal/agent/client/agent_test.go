package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrPTqp/metrics/internal/models"
)

func TestSendMetrics(t *testing.T) {
	tests := []struct {
		name           string
		gauges         map[string]float64
		counters       map[string]int64
		serverStatus   int
		serverResponse string
		wantErr        bool
		expectRequest  bool
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
		},
		{
			name:           "empty metrics - no request should be sent",
			gauges:         map[string]float64{},
			counters:       map[string]int64{},
			serverStatus:   http.StatusOK,
			serverResponse: "",
			wantErr:        false,
			expectRequest:  false,
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
					t.Errorf("Failed to read request body: %v", err)
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
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			serverURL := server.URL[7:] // убираем "http://"

			client := &http.Client{}

			err := sendMetrics(tt.gauges, tt.counters, client, serverURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("sendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}

			if requestReceived != tt.expectRequest {
				t.Errorf("Request received = %v, expectRequest %v", requestReceived, tt.expectRequest)
			}

			if requestReceived && tt.serverStatus == http.StatusOK {
				if contentType := requestHeaders.Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Content-Type header = %s, expected application/json", contentType)
				}
				if contentEncoding := requestHeaders.Get("Content-Encoding"); contentEncoding != "gzip" {
					t.Errorf("Content-Encoding header = %s, expected gzip", contentEncoding)
				}

				if len(requestBody) > 0 {
					decompressedBody, err := Decompress(requestBody)
					if err != nil {
						t.Errorf("Failed to decompress request body: %v", err)
					} else {
						var metrics []models.Metrics
						if err := json.Unmarshal(decompressedBody, &metrics); err != nil {
							t.Errorf("Invalid JSON in request: %v", err)
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
				}
			}
		})
	}
}

func TestSendMetrics_RequestStructure(t *testing.T) {
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
				json.Unmarshal(decompressedBody, &receivedMetrics)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := &http.Client{}
			err := sendMetrics(tt.gauges, tt.counters, client, server.URL[7:])

			if err != nil {
				t.Fatalf("sendMetrics() unexpected error: %v", err)
			}

			if len(receivedMetrics) != len(tt.gauges)+len(tt.counters) {
				t.Errorf("Expected %d metrics, got %d", len(tt.gauges)+len(tt.counters), len(receivedMetrics))
			}

			for name, expectedValue := range tt.gauges {
				found := false
				for _, metric := range receivedMetrics {
					if metric.ID == name && metric.MType == "gauge" && metric.Value != nil {
						if *metric.Value != expectedValue {
							t.Errorf("Gauge %s value = %f, expected %f", name, *metric.Value, expectedValue)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Gauge metric %s not found in request", name)
				}
			}

			for name, expectedDelta := range tt.counters {
				found := false
				for _, metric := range receivedMetrics {
					if metric.ID == name && metric.MType == "counter" && metric.Delta != nil {
						if *metric.Delta != expectedDelta {
							t.Errorf("Counter %s delta = %d, expected %d", name, *metric.Delta, expectedDelta)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Counter metric %s not found in request", name)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
