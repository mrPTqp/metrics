package repository

type MetricRepository interface {
	SaveGauge(name string, value *float64) error
    SaveCounter(name string, value *int64) error
    GetGauge(name string) (float64, error)
    GetCounter(name string) (int64, error)
    ListGauges() (map[string]float64, error)
    ListCounters() (map[string]int64, error) 
    SaveAllMetrics(map[string]float64, map[string]int64) error
}
