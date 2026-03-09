package backup

import (
	"context"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
)

type stubMetricsLister struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (s *stubMetricsLister) ListAllMetrics(context.Context) (map[string]float64, map[string]int64) {
	return s.gauges, s.counters
}

var _ service.MetricsLister = (*stubMetricsLister)(nil)

func TestBackuper_Backup(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "snapshot.json")
	logger := zaptest.NewLogger(t)

	producer := storage.NewFileProducer(filePath, logger)
	consumer := storage.NewFileConsumer(filePath, logger)
	fs := storage.NewFileStorage(producer, consumer, nil, false, logger)

	lister := &stubMetricsLister{
		gauges:   map[string]float64{"g1": 1},
		counters: map[string]int64{"c1": 2},
	}

	b := NewBackuper(nil, lister, fs, logger)

	ctx := contextkey.WithLogger(context.Background(), logger)
	if err := b.Backup(ctx); err != nil {
		t.Fatalf("Backup() error = %v, want nil", err)
	}

	data, err := consumer.ReadData()
	if err != nil {
		t.Fatalf("ReadData() error = %v, want nil", err)
	}
	if data.Gauges["g1"] != 1 {
		t.Errorf("backup gauges[\"g1\"] = %v, want %v", data.Gauges["g1"], 1.0)
	}
	if data.Counters["c1"] != 2 {
		t.Errorf("backup counters[\"c1\"] = %v, want %v", data.Counters["c1"], 2)
	}
}

