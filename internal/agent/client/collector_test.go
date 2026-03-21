package agent

import "testing"

func TestCollectGaugeMetrics_NotNilAndContainsBasicKeys(t *testing.T) {
	metrics := CollectGaugeMetrics()
	if metrics == nil {
		t.Fatal("CollectGaugeMetrics() returned nil map")
	}
	// карта должна содержать хотя бы некоторые базовые ключи из runtime.MemStats
	requiredKeys := []string{
		"GCCPUFraction",
		"HeapAlloc",
		"Sys",
	}
	for _, k := range requiredKeys {
		if _, ok := metrics[k]; !ok {
			t.Errorf("CollectGaugeMetrics() missing key %q", k)
		}
	}
}

func TestCollectAdditionalGaugeMetrics_NotNil(t *testing.T) {
	metrics := CollectAdditionalGaugeMetrics()
	if metrics == nil {
		t.Fatal("CollectAdditionalGaugeMetrics() returned nil map")
	}
	// gopsutil может вернуть ошибку в некоторых средах, поэтому не проверяем наличие конкретных ключей,
	// а только то, что функция не падает и всегда возвращает непустую карту (возможно, пустую по содержимому).
}

