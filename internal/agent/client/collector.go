package agent

import "runtime"

func CollectMetrics() map[string]float64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := make(map[string]float64)
	metrics["GCCPUFraction"] = memStats.GCCPUFraction
	metrics["GCSys"] = float64(memStats.GCSys)
	metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	metrics["HeapIdle"] = float64(memStats.HeapIdle)
	metrics["HeapInuse"] = float64(memStats.HeapInuse)
	metrics["HeapObjects"] = float64(memStats.HeapObjects)
	metrics["HeapReleased"] = float64(memStats.HeapReleased)
	metrics["HeapSys"] = float64(memStats.HeapSys)
	metrics["LastGC"] = float64(memStats.LastGC)
	metrics["Lookups"] = float64(memStats.Lookups)
	metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	metrics["MCacheSys"] = float64(memStats.MCacheSys)
	metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	metrics["MSpanSys"] = float64(memStats.MSpanSys)
	metrics["Mallocs"] = float64(memStats.Mallocs)
	metrics["NextGC"] = float64(memStats.NextGC)
	metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	metrics["NumGC"] = float64(memStats.NumGC)
	metrics["OtherSys"] = float64(memStats.OtherSys)
	metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	metrics["StackInuse"] = float64(memStats.StackInuse)
	metrics["StackSys"] = float64(memStats.StackSys)
	metrics["Sys"] = float64(memStats.Sys)
	metrics["TotalAlloc"] = float64(memStats.TotalAlloc)
	metrics["Alloc"] = float64(memStats.Alloc)
	metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	metrics["Frees"] = float64(memStats.Frees)
	return metrics
}
