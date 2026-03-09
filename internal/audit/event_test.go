package audit

import (
	"context"
	"encoding/json"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAuditEvent_MarshalWithNewline(t *testing.T) {
	// Параметризованный тест
	tests := []struct {
		name  string
		event AuditEvent
	}{
		{
			name: "valid event with metrics",
			event: AuditEvent{
				TS:        time.Now().Unix(),
				Metrics:   []string{"cpu_usage", "memory_usage"},
				IPAddress: "192.168.1.1",
			},
		},
		{
			name: "event with empty metrics",
			event: AuditEvent{
				TS:        time.Now().Unix(),
				Metrics:   []string{},
				IPAddress: "192.168.1.2",
			},
		},
		{
			name: "event with empty IP",
			event: AuditEvent{
				TS:        time.Now().Unix(),
				Metrics:   []string{"disk_usage"},
				IPAddress: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.event.marshalWithNewline()

			if data[len(data)-1] != '\n' {
				t.Error("marshaled data does not end with newline")
			}


			var unmarshaledEvent AuditEvent
			if err := json.Unmarshal(data[:len(data)-1], &unmarshaledEvent); err != nil {
				t.Fatalf("Failed to unmarshal event: %v", err)
			}

			if unmarshaledEvent.TS != tt.event.TS {
				t.Errorf("Event TS = %v, want %v", unmarshaledEvent.TS, tt.event.TS)
			}
			if len(unmarshaledEvent.Metrics) != len(tt.event.Metrics) {
				t.Errorf("Event Metrics count = %v, want %v", len(unmarshaledEvent.Metrics), len(tt.event.Metrics))
			}
			for i, m := range unmarshaledEvent.Metrics {
				if m != tt.event.Metrics[i] {
					t.Errorf("Event Metrics[%d] = %v, want %v", i, m, tt.event.Metrics[i])
				}
			}
			if unmarshaledEvent.IPAddress != tt.event.IPAddress {
				t.Errorf("Event IPAddress = %v, want %v", unmarshaledEvent.IPAddress, tt.event.IPAddress)
			}
		})
	}
}

func TestAuditEvent_JSONMarshal(t *testing.T) {
	// Проверка корректности JSON-сериализации
	event := AuditEvent{
		TS:        1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "192.168.1.1",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	expected := `{"ts":1234567890,"metrics":["metric1","metric2"],"ip_address":"192.168.1.1"}`
	if string(data) != expected {
		t.Errorf("JSON output = %s, want %s", string(data), expected)
	}
}

type fakeAuditProcessor struct {
	ch chan AuditEvent
}

func newFakeAuditProcessor() *fakeAuditProcessor {
	return &fakeAuditProcessor{
		ch: make(chan AuditEvent, 10),
	}
}

func (f *fakeAuditProcessor) Write(e AuditEvent) error {
	f.ch <- e
	return nil
}

func TestEventBus_PublishAndProcess(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name   string
		events []AuditEvent
	}{
		{
			name: "single event",
			events: []AuditEvent{
				{
					TS:        time.Now().Unix(),
					Metrics:   []string{"cpu", "ram"},
					IPAddress: "10.0.0.1",
				},
			},
		},
		{
			name: "multiple events",
			events: []AuditEvent{
				{
					TS:        time.Now().Unix(),
					Metrics:   []string{"metric1"},
					IPAddress: "127.0.0.1",
				},
				{
					TS:        time.Now().Unix() + 1,
					Metrics:   []string{"metric2", "metric3"},
					IPAddress: "192.168.0.10",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := newFakeAuditProcessor()
			bus := NewEventBus([]AuditProcessor{processor}, 10, logger)
			for _, e := range tt.events {
				if err := bus.Publish(e); err != nil {
					t.Fatalf("Publish returned error: %v", err)
				}

	
				deadline := time.Now().Add(2 * time.Second)
				for {
					select {
					case got := <-processor.ch:
						if got.TS != e.TS || got.IPAddress != e.IPAddress {
							t.Fatalf("processor received wrong event: %+v, want %+v", got, e)
						}
						goto processed
					default:
					}
					if time.Now().After(deadline) {
						t.Fatal("processor did not receive event in time")
					}
					time.Sleep(10 * time.Millisecond)
				}
			processed:
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := bus.ShutDown(ctx); err != nil {
				t.Fatalf("ShutDown returned error: %v", err)
			}
		})
	}
}

func TestEventBus_PublishFull(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
	}{
		{
			name:       "buffer size 1",
			bufferSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bus := &EventBus{
				events: make(chan AuditEvent, tt.bufferSize),
			}

			bus.events <- AuditEvent{}
			if err := bus.Publish(AuditEvent{}); err == nil {
				t.Fatal("expected error when publishing to full bus, got nil")
			}
		})
	}
}

func TestEventBus_ShutdownContextCanceled(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "context canceled before shutdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := &EventBus{
				events: make(chan AuditEvent),
				done:   make(chan struct{}), // никогда не закрывается
				jobs:   make(chan job),
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			if err := bus.ShutDown(ctx); err == nil {
				t.Fatal("expected context error from ShutDown, got nil")
			}
		})
	}
}

func TestFileAuditProcessor_WriteAndShutdown(t *testing.T) {
	tests := []struct {
		name  string
		event AuditEvent
	}{
		{
			name: "single event written via channel",
			event: AuditEvent{
				TS:        42,
				Metrics:   []string{"m1", "m2"},
				IPAddress: "127.0.0.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "audit-*.log")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			tmpName := tmpFile.Name()
			_ = tmpFile.Close()
			defer os.Remove(tmpName)

			logger := zap.NewNop()
			processor, err := NewFileAuditProcessor(tmpName, logger)
			if err != nil {
				t.Fatalf("failed to create FileAuditProcessor: %v", err)
			}


			_ = processor.Write(tt.event)


			processor.ch <- tt.event.marshalWithNewline()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := processor.ShutDown(ctx); err != nil {
				t.Fatalf("ShutDown returned error: %v", err)
			}

			data, err := os.ReadFile(tmpName)
			if err != nil {
				t.Fatalf("failed to read audit file: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("audit file is empty, expected written event")
			}

			lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
			var got AuditEvent
			if err := json.Unmarshal(lines[0], &got); err != nil {
				t.Fatalf("failed to unmarshal written event: %v", err)
			}
			if got.TS != tt.event.TS || got.IPAddress != tt.event.IPAddress {
				t.Fatalf("written event mismatch: got %+v, want %+v", got, tt.event)
			}
		})
	}
}

func TestHTTPAuditProcessor_WriteAndShutdown(t *testing.T) {
	tests := []struct {
		name  string
		event AuditEvent
	}{
		{
			name: "basic write and shutdown",
			event: AuditEvent{
				TS:        1,
				Metrics:   []string{"a"},
				IPAddress: "1.2.3.4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&count, 1)
				defer r.Body.Close()
			}))
			defer server.Close()

			logger := zap.NewNop()
			processor := NewHTTPAuditProcessor(server.URL, logger)


			processor.client = server.Client()


			processor.send(tt.event)

			if atomic.LoadInt32(&count) == 0 {
				t.Fatal("expected at least one HTTP request from send, got 0")
			}

			if err := processor.Write(tt.event); err != nil {
				t.Fatalf("Write returned error: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := processor.ShutDown(ctx); err != nil {
				t.Fatalf("ShutDown returned error: %v", err)
			}
		})
	}
}

func TestHTTPAuditProcessor_ShutdownContextCanceled(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "context canceled before shutdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &HTTPAuditProcessor{
				ch:     make(chan AuditEvent),
				done:   make(chan struct{}), // никогда не закрывается
				client: &http.Client{},
				URL:    "http://example.com",
				logger: zap.NewNop(),
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			if err := processor.ShutDown(ctx); err == nil {
				t.Fatal("expected context error from ShutDown, got nil")
			}
		})
	}
}

