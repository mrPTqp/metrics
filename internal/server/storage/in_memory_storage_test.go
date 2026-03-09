package storage

import (
	"context"
	"errors"
	"testing"
)

func TestMemStorage_GaugeAndCounter(t *testing.T) {
	s := NewMemStorage()
	ctx := context.Background()

	t.Run("save and get gauge", func(t *testing.T) {
		value := 1.23
		if err := s.SaveGauge(ctx, "test_gauge", &value); err != nil {
			t.Fatalf("SaveGauge() error = %v, want nil", err)
		}

		got, err := s.GetGauge(ctx, "test_gauge")
		if err != nil {
			t.Fatalf("GetGauge() error = %v, want nil", err)
		}
		if got != value {
			t.Errorf("GetGauge() = %v, want %v", got, value)
		}
	})

	t.Run("gauge not found", func(t *testing.T) {
		if _, err := s.GetGauge(ctx, "unknown"); !errors.Is(err, ErrMetricNotFound) {
			t.Fatalf("GetGauge() error = %v, want %v", err, ErrMetricNotFound)
		}
	})

	t.Run("save and get counter", func(t *testing.T) {
		v1 := int64(10)
		if err := s.SaveCounter(ctx, "cnt", &v1); err != nil {
			t.Fatalf("SaveCounter() error = %v, want nil", err)
		}

		v2 := int64(5)
		if err := s.SaveCounter(ctx, "cnt", &v2); err != nil {
			t.Fatalf("SaveCounter() error = %v, want nil", err)
		}

		got, err := s.GetCounter(ctx, "cnt")
		if err != nil {
			t.Fatalf("GetCounter() error = %v, want nil", err)
		}
		if want := v1 + v2; got != want {
			t.Errorf("GetCounter() = %d, want %d", got, want)
		}
	})
}

func TestMemStorage_ListAndSaveAllMetrics(t *testing.T) {
	s := NewMemStorage()
	ctx := context.Background()

	t.Run("list gauges and counters and ensure copy", func(t *testing.T) {
		v := 3.14
		if err := s.SaveGauge(ctx, "g1", &v); err != nil {
			t.Fatalf("SaveGauge() error = %v", err)
		}
		c := int64(7)
		if err := s.SaveCounter(ctx, "c1", &c); err != nil {
			t.Fatalf("SaveCounter() error = %v", err)
		}

		gauges, err := s.ListGauges(ctx)
		if err != nil {
			t.Fatalf("ListGauges() error = %v", err)
		}
		counters, err := s.ListCounters(ctx)
		if err != nil {
			t.Fatalf("ListCounters() error = %v", err)
		}

		if gauges["g1"] != v {
			t.Errorf("ListGauges()[\"g1\"] = %v, want %v", gauges["g1"], v)
		}
		if counters["c1"] != c {
			t.Errorf("ListCounters()[\"c1\"] = %v, want %v", counters["c1"], c)
		}

		gauges["g1"] = 0
		counters["c1"] = 0
		if orig, _ := s.GetGauge(ctx, "g1"); orig != v {
			t.Errorf("underlying gauge modified via returned map: %v, want %v", orig, v)
		}
		if orig, _ := s.GetCounter(ctx, "c1"); orig != c {
			t.Errorf("underlying counter modified via returned map: %v, want %v", orig, c)
		}
	})

	t.Run("SaveAllMetrics and reset with nil", func(t *testing.T) {
		gauges := map[string]float64{"g1": 1.0}
		counters := map[string]int64{"c1": 2}

		if err := s.SaveAllMetrics(ctx, gauges, counters); err != nil {
			t.Fatalf("SaveAllMetrics() error = %v, want nil", err)
		}

		if got, _ := s.GetGauge(ctx, "g1"); got != 1.0 {
			t.Errorf("GetGauge() after SaveAllMetrics = %v, want %v", got, 1.0)
		}
		if got, _ := s.GetCounter(ctx, "c1"); got != 2 {
			t.Errorf("GetCounter() after SaveAllMetrics = %v, want %v", got, 2)
		}

		if err := s.SaveAllMetrics(ctx, nil, nil); err != nil {
			t.Fatalf("SaveAllMetrics() with nil maps error = %v", err)
		}
		if _, err := s.GetGauge(ctx, "g1"); !errors.Is(err, ErrMetricNotFound) {
			t.Fatalf("GetGauge() after reset error = %v, want %v", err, ErrMetricNotFound)
		}
		if _, err := s.GetCounter(ctx, "c1"); !errors.Is(err, ErrMetricNotFound) {
			t.Fatalf("GetCounter() after reset error = %v, want %v", err, ErrMetricNotFound)
		}
	})
}

func TestMemStorage_Others(t *testing.T) {
	s := NewMemStorage()
	if !s.CheckStorageAvailability(context.Background()) {
		t.Error("CheckStorageAvailability() = false, want true")
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}


