package storage

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestFileProducerAndConsumer(t *testing.T) {
	t.Run("write and read snapshot data", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "snapshot.json")
		logger := zaptest.NewLogger(t)

		producer := NewFileProducer(filePath, logger)
		consumer := NewFileConsumer(filePath, logger)

		orig := snapshotData{
			Gauges:   map[string]float64{"g1": 1.23},
			Counters: map[string]int64{"c1": 42},
		}

		if err := producer.WriteData(orig); err != nil {
			t.Fatalf("WriteData() error = %v, want nil", err)
		}

		read, err := consumer.ReadData()
		if err != nil {
			t.Fatalf("ReadData() error = %v, want nil", err)
		}

		if read.Gauges["g1"] != orig.Gauges["g1"] {
			t.Errorf("ReadData().Gauges[\"g1\"] = %v, want %v", read.Gauges["g1"], orig.Gauges["g1"])
		}
		if read.Counters["c1"] != orig.Counters["c1"] {
			t.Errorf("ReadData().Counters[\"c1\"] = %v, want %v", read.Counters["c1"], orig.Counters["c1"])
		}
	})

	t.Run("read data from missing file", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "missing.json")
		logger := zaptest.NewLogger(t)

		consumer := NewFileConsumer(filePath, logger)
		data, err := consumer.ReadData()
		if err != nil {
			t.Fatalf("ReadData() error = %v, want nil", err)
		}
		if len(data.Gauges) != 0 || len(data.Counters) != 0 {
			t.Errorf("ReadData() from missing file = %#v, want empty maps", data)
		}
	})

	t.Run("read data from empty file", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "snapshot.json")
		logger := zaptest.NewLogger(t)

		if err := os.WriteFile(filePath, []byte(""), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		consumer := NewFileConsumer(filePath, logger)
		data, err := consumer.ReadData()
		if err != nil {
			t.Fatalf("ReadData() error = %v, want nil", err)
		}
		if len(data.Gauges) != 0 || len(data.Counters) != 0 {
			t.Errorf("ReadData() from empty file = %#v, want empty maps", data)
		}
	})
}

func TestFileConsumer_CheckFileAccess(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.json")
	missing := filepath.Join(dir, "missing.json")
	logger := zaptest.NewLogger(t)

	if err := os.WriteFile(existing, []byte("{}"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	consumerExisting := NewFileConsumer(existing, logger)
	if !consumerExisting.CheckFileAccess() {
		t.Error("CheckFileAccess() for existing file = false, want true")
	}

	consumerMissing := NewFileConsumer(missing, logger)
	if consumerMissing.CheckFileAccess() {
		t.Error("CheckFileAccess() for missing file = true, want false")
	}
}


