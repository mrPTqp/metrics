package handler

type mockMetricsService struct {
	saveGaugeFunc      func(name string, value *float64) error
	saveCounterFunc    func(name string, value *int64) error
	getGaugeFunc       func(name string) (float64, error)
	getCounterFunc     func(name string) (int64, error)
	listAllFunc        func() (map[string]float64, map[string]int64)
	saveAllMetricsFunc func(gauges map[string]float64, counters map[string]int64) error
	pingFunc           func() bool
}

func (m *mockMetricsService) SaveGaugeMetric(name string, value *float64) error {
	if m.saveGaugeFunc != nil {
		return m.saveGaugeFunc(name, value)
	}
	return nil
}

func (m *mockMetricsService) SaveCounterMetric(name string, value *int64) error {
	if m.saveCounterFunc != nil {
		return m.saveCounterFunc(name, value)
	}
	return nil
}

func (m *mockMetricsService) GetGaugeMetric(name string) (float64, error) {
	if m.getGaugeFunc != nil {
		return m.getGaugeFunc(name)
	}
	return 0, nil
}

func (m *mockMetricsService) GetCounterMetric(name string) (int64, error) {
	if m.getCounterFunc != nil {
		return m.getCounterFunc(name)
	}
	return 0, nil
}

func (m *mockMetricsService) ListAllMetrics() (map[string]float64, map[string]int64) {
	if m.listAllFunc != nil {
		return m.listAllFunc()
	}
	return map[string]float64{}, map[string]int64{}
}

func (m *mockMetricsService) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	if m.saveAllMetricsFunc != nil {
		return m.saveAllMetricsFunc(gauges, counters)
	}
	return nil
}

func (m *mockMetricsService) Ping() bool {
	if m.pingFunc != nil {
		return m.pingFunc()
	}
	return true
}