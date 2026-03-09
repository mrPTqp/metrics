package models

import "testing"

func TestMetrics_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
		m    Metrics
	}{
		{
			name: "zero value struct",
			m:    Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			if m.ID != "" {
				t.Errorf("ID = %q, want empty", m.ID)
			}
			if m.MType != "" {
				t.Errorf("MType = %q, want empty", m.MType)
			}
			if m.Delta != nil {
				t.Errorf("Delta = %v, want nil", m.Delta)
			}
			if m.Value != nil {
				t.Errorf("Value = %v, want nil", m.Value)
			}
			if m.Hash != "" {
				t.Errorf("Hash = %q, want empty", m.Hash)
			}
		})
	}
}

func TestMetrics_TypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{name: "Counter", got: Counter, expected: "counter"},
		{name: "Gauge", got: Gauge, expected: "gauge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s constant = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}


