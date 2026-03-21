package storage

import (
	"context"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"
)

type stubMetricsProvider struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (s *stubMetricsProvider) ListGauges(context.Context) (map[string]float64, error) {
	return s.gauges, nil
}

func (s *stubMetricsProvider) ListCounters(context.Context) (map[string]int64, error) {
	return s.counters, nil
}

func TestFileStorage_BackupAndRestore(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "snapshot.json")
	logger := zaptest.NewLogger(t)

	producer := NewFileProducer(filePath, logger)
	consumer := NewFileConsumer(filePath, logger)

	mp := &stubMetricsProvider{
		gauges:   map[string]float64{"g1": 1.0},
		counters: map[string]int64{"c1": 2},
	}

	fs := NewFileStorage(producer, consumer, mp, true, logger)

	if err := fs.Backup(); err != nil {
		t.Fatalf("Backup() error = %v, want nil", err)
	}

	gauges, counters, err := fs.Restore()
	if err != nil {
		t.Fatalf("Restore() error = %v, want nil", err)
	}

	if gauges["g1"] != 1.0 {
		t.Errorf("Restore() gauges[\"g1\"] = %v, want %v", gauges["g1"], 1.0)
	}
	if counters["c1"] != 2 {
		t.Errorf("Restore() counters[\"c1\"] = %v, want %v", counters["c1"], 2)
	}
}

func TestFileStorage_SaveAllMetrics(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "snapshot.json")
	logger := zaptest.NewLogger(t)

	producer := NewFileProducer(filePath, logger)
	consumer := NewFileConsumer(filePath, logger)
	mp := &stubMetricsProvider{}

	fs := NewFileStorage(producer, consumer, mp, false, logger)

	gauges := map[string]float64{"g1": 3.14}
	counters := map[string]int64{"c1": 7}

	if err := fs.SaveAllMetrics(context.Background(), gauges, counters); err != nil {
		t.Fatalf("SaveAllMetrics() error = %v, want nil", err)
	}

	data, err := consumer.ReadData()
	if err != nil {
		t.Fatalf("ReadData() error = %v, want nil", err)
	}
	if data.Gauges["g1"] != 3.14 {
		t.Errorf("saved gauges[\"g1\"] = %v, want %v", data.Gauges["g1"], 3.14)
	}
	if data.Counters["c1"] != 7 {
		t.Errorf("saved counters[\"c1\"] = %v, want %v", data.Counters["c1"], 7)
	}
}

func TestFileStorage_SaveGaugeAndCounter_TriggerBackupWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "snapshot.json")
	logger := zaptest.NewLogger(t)

	producer := NewFileProducer(filePath, logger)
	consumer := NewFileConsumer(filePath, logger)

	mp := &stubMetricsProvider{
		gauges:   map[string]float64{"g1": 1},
		counters: map[string]int64{"c1": 2},
	}

	fs := NewFileStorage(producer, consumer, mp, true, logger)

	ctx := context.Background()
	gVal := 5.0
	if err := fs.SaveGauge(ctx, "g1", &gVal); err != nil {
		t.Fatalf("SaveGauge() error = %v, want nil", err)
	}
	cVal := int64(10)
	if err := fs.SaveCounter(ctx, "c1", &cVal); err != nil {
		t.Fatalf("SaveCounter() error = %v, want nil", err)
	}

	data, err := consumer.ReadData()
	if err != nil {
		t.Fatalf("ReadData() error = %v, want nil", err)
	}
	if len(data.Gauges) == 0 || len(data.Counters) == 0 {
		t.Errorf("expected backup snapshot to be written, got %#v", data)
	}
}

func TestFileStorage_CheckStorageAvailabilityAndClose(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "snapshot.json")
	logger := zaptest.NewLogger(t)

	producer := NewFileProducer(filePath, logger)
	consumer := NewFileConsumer(filePath, logger)
	mp := &stubMetricsProvider{}

	fs := NewFileStorage(producer, consumer, mp, false, logger)

	if ok := fs.CheckStorageAvailability(context.Background()); ok {
		// file does not exist yet, should be false
		t.Errorf("CheckStorageAvailability() = true for non-existent file, want false")
	}

	// create file to make it accessible
	if err := producer.WriteData(snapshotData{}); err != nil {
		t.Fatalf("WriteData() error = %v", err)
	}

	if ok := fs.CheckStorageAvailability(context.Background()); !ok {
		t.Errorf("CheckStorageAvailability() = false for existing file, want true")
	}

	if err := fs.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

