package service

type MetricsService interface {
	SaveGaugeMetric(mName string, mValue *float64) error
	SaveCounterMetric(mName string, mValue *int64) error
	GetGaugeMetric(mName string) (float64, error)
	GetCounterMetric(mName string) (int64, error)
	ListAllMetrics() (map[string]float64, map[string]int64)
	SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error
	Ping() bool
}
