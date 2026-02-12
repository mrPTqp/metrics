package repository

// Репозиторий для хранения метрик
type MetricRepository interface {
	SaveAllMetrics(gauges map[string]float64, counters map[string]int64)
	GetAllMetrics() (map[string]float64, map[string]int64)
	SaveAdditionalGaugeMetrics(additionalGauges map[string]float64)
	GetAdditionalGaugeMetrics() map[string]float64
}
