package storage

import "testing"

func TestMemStorage_SaveAndGetAllMetrics(t *testing.T) {
	s := NewMemStorage()

	gauges := map[string]float64{"g1": 1.0}
	counters := map[string]int64{"c1": 2}
	s.SaveAllMetrics(gauges, counters)

	gotG, gotC := s.GetAllMetrics()
	if gotG["g1"] != 1.0 {
		t.Errorf("GetAllMetrics gauges[\"g1\"] = %v, want %v", gotG["g1"], 1.0)
	}
	if gotC["c1"] != 2 {
		t.Errorf("GetAllMetrics counters[\"c1\"] = %v, want %v", gotC["c1"], 2)
	}

	// ensure returned maps are copies
	gotG["g1"] = 0
	gotC["c1"] = 0
	origG, origC := s.GetAllMetrics()
	if origG["g1"] != 1.0 || origC["c1"] != 2 {
		t.Errorf("underlying maps modified via returned maps")
	}
}

func TestMemStorage_SaveAndGetAdditionalGaugeMetrics(t *testing.T) {
	s := NewMemStorage()

	additional := map[string]float64{"a1": 3.14}
	s.SaveAdditionalGaugeMetrics(additional)

	got := s.GetAdditionalGaugeMetrics()
	if got["a1"] != 3.14 {
		t.Errorf("GetAdditionalGaugeMetrics()[\"a1\"] = %v, want %v", got["a1"], 3.14)
	}

	// nil should reset
	s.SaveAdditionalGaugeMetrics(nil)
	got = s.GetAdditionalGaugeMetrics()
	if len(got) != 0 {
		t.Errorf("expected empty additional gauges after reset, got %#v", got)
	}
}

