package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
)

type stubMetricWriter struct {
	lastGauges   map[string]float64
	lastCounters map[string]int64
	callCount    int
	err          error
}

func (s *stubMetricWriter) SaveGaugeMetric(context.Context, string, *float64) error {
	return nil
}
func (s *stubMetricWriter) SaveCounterMetric(context.Context, string, *int64) error {
	return nil
}
func (s *stubMetricWriter) SaveAllMetrics(_ context.Context, gauges map[string]float64, counters map[string]int64) error {
	s.callCount++
	s.lastGauges = gauges
	s.lastCounters = counters
	return s.err
}

var _ service.MetricWriter = (*stubMetricWriter)(nil)

func TestRestorer_Restore_WithMetrics(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "snapshot.json")
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		jsonData       string
		wantCalls      int
		wantGaugeValue float64
		wantCounterVal int64
	}{
		{
			name:           "with metrics",
			jsonData:       `{"gauges":{"g1":1},"counters":{"c1":2}}`,
			wantCalls:      1,
			wantGaugeValue: 1,
			wantCounterVal: 2,
		},
		{
			name:      "empty snapshot",
			jsonData:  `{"gauges":{},"counters":{}}`,
			wantCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(filePath, []byte(tt.jsonData), 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			consumer := storage.NewFileConsumer(filePath, logger)
			fs := storage.NewFileStorage(nil, consumer, nil, false, logger)

			writer := &stubMetricWriter{}
			r := NewRestorer(writer, fs, logger)

			ctx := contextkey.WithLogger(context.Background(), logger)
			if err := r.Restore(ctx); err != nil {
				t.Fatalf("Restore() error = %v, want nil", err)
			}

			if writer.callCount != tt.wantCalls {
				t.Fatalf("SaveAllMetrics called %d times, want %d", writer.callCount, tt.wantCalls)
			}
			if tt.wantCalls > 0 {
				if writer.lastGauges["g1"] != tt.wantGaugeValue {
					t.Errorf("restored gauges[\"g1\"] = %v, want %v", writer.lastGauges["g1"], tt.wantGaugeValue)
				}
				if writer.lastCounters["c1"] != tt.wantCounterVal {
					t.Errorf("restored counters[\"c1\"] = %v, want %v", writer.lastCounters["c1"], tt.wantCounterVal)
				}
			}
		})
	}
}

