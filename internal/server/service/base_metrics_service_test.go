package service

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/contextkey"
)

type stubMetricRepository struct {
	gauges          map[string]float64
	counters        map[string]int64
	saveGaugeErr    error
	saveCounterErr  error
	listGaugesErr   error
	listCountersErr error
	pingResult      bool
}

func (s *stubMetricRepository) SaveGauge(_ context.Context, name string, value *float64) error {
	if s.saveGaugeErr != nil {
		return s.saveGaugeErr
	}
	if s.gauges == nil {
		s.gauges = make(map[string]float64)
	}
	if value != nil {
		s.gauges[name] = *value
	}
	return nil
}

func (s *stubMetricRepository) SaveCounter(_ context.Context, name string, value *int64) error {
	if s.saveCounterErr != nil {
		return s.saveCounterErr
	}
	if s.counters == nil {
		s.counters = make(map[string]int64)
	}
	if value != nil {
		s.counters[name] += *value
	}
	return nil
}

func (s *stubMetricRepository) GetGauge(_ context.Context, name string) (float64, error) {
	if v, ok := s.gauges[name]; ok {
		return v, nil
	}
	return 0, errors.New("not found")
}

func (s *stubMetricRepository) GetCounter(_ context.Context, name string) (int64, error) {
	if v, ok := s.counters[name]; ok {
		return v, nil
	}
	return 0, errors.New("not found")
}

func (s *stubMetricRepository) ListGauges(_ context.Context) (map[string]float64, error) {
	if s.listGaugesErr != nil {
		return nil, s.listGaugesErr
	}
	if s.gauges == nil {
		return make(map[string]float64), nil
	}
	return s.gauges, nil
}

func (s *stubMetricRepository) ListCounters(_ context.Context) (map[string]int64, error) {
	if s.listCountersErr != nil {
		return nil, s.listCountersErr
	}
	if s.counters == nil {
		return make(map[string]int64), nil
	}
	return s.counters, nil
}

func (s *stubMetricRepository) SaveAllMetrics(_ context.Context, gauges map[string]float64, counters map[string]int64) error {
	s.gauges = gauges
	s.counters = counters
	return nil
}

func (s *stubMetricRepository) CheckStorageAvailability(_ context.Context) bool {
	return s.pingResult
}

func (s *stubMetricRepository) Close() error {
	return nil
}

func newTestService(t *testing.T, repo *stubMetricRepository) *BaseMetricService {
	t.Helper()
	logger := zaptest.NewLogger(t)
	return NewMetricsService(repo, logger)
}

func newCtxWithLogger(t *testing.T) context.Context {
	t.Helper()
	return contextkey.WithLogger(context.Background(), zaptest.NewLogger(t))
}

func TestBaseMetricService_SaveGaugeMetric(t *testing.T) {
	repo := &stubMetricRepository{}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	value := 3.5
	if err := svc.SaveGaugeMetric(ctx, "g1", &value); err != nil {
		t.Fatalf("SaveGaugeMetric() error = %v, want nil", err)
	}
	if repo.gauges["g1"] != value {
		t.Errorf("stored gauge = %v, want %v", repo.gauges["g1"], value)
	}
}

func TestBaseMetricService_SaveGaugeMetric_Error(t *testing.T) {
	wantErr := errors.New("save failed")
	repo := &stubMetricRepository{saveGaugeErr: wantErr}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	value := 1.0
	if err := svc.SaveGaugeMetric(ctx, "g1", &value); !errors.Is(err, wantErr) {
		t.Fatalf("SaveGaugeMetric() error = %v, want %v", err, wantErr)
	}
}

func TestBaseMetricService_SaveCounterMetric(t *testing.T) {
	repo := &stubMetricRepository{}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	value := int64(5)
	if err := svc.SaveCounterMetric(ctx, "c1", &value); err != nil {
		t.Fatalf("SaveCounterMetric() error = %v, want nil", err)
	}
	if repo.counters["c1"] != value {
		t.Errorf("stored counter = %v, want %v", repo.counters["c1"], value)
	}
}

func TestBaseMetricService_GetGaugeMetric(t *testing.T) {
	repo := &stubMetricRepository{
		gauges: map[string]float64{"g1": 2.5},
	}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	got, err := svc.GetGaugeMetric(ctx, "g1")
	if err != nil {
		t.Fatalf("GetGaugeMetric() error = %v, want nil", err)
	}
	if got != 2.5 {
		t.Errorf("GetGaugeMetric() = %v, want %v", got, 2.5)
	}
}

func TestBaseMetricService_ListAllMetrics(t *testing.T) {
	repo := &stubMetricRepository{
		gauges:   map[string]float64{"g1": 1},
		counters: map[string]int64{"c1": 2},
	}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	gauges, counters := svc.ListAllMetrics(ctx)
	if gauges["g1"] != 1 {
		t.Errorf("ListAllMetrics gauges[\"g1\"] = %v, want %v", gauges["g1"], 1.0)
	}
	if counters["c1"] != 2 {
		t.Errorf("ListAllMetrics counters[\"c1\"] = %v, want %v", counters["c1"], 2)
	}
}

func TestBaseMetricService_ListAllMetrics_Errors(t *testing.T) {
	repo := &stubMetricRepository{
		listGaugesErr:   errors.New("gauges error"),
		listCountersErr: errors.New("counters error"),
	}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	gauges, counters := svc.ListAllMetrics(ctx)
	if len(gauges) != 0 {
		t.Errorf("gauges len = %d, want 0 on error", len(gauges))
	}
	if len(counters) != 0 {
		t.Errorf("counters len = %d, want 0 on error", len(counters))
	}
}

func TestBaseMetricService_SaveAllMetricsAndPing(t *testing.T) {
	repo := &stubMetricRepository{pingResult: true}
	svc := newTestService(t, repo)
	ctx := newCtxWithLogger(t)

	gauges := map[string]float64{"g1": 1}
	counters := map[string]int64{"c1": 2}
	if err := svc.SaveAllMetrics(ctx, gauges, counters); err != nil {
		t.Fatalf("SaveAllMetrics() error = %v, want nil", err)
	}
	if ok := svc.Ping(ctx); !ok {
		t.Errorf("Ping() = false, want true")
	}
}

