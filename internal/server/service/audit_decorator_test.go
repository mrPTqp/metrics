package service

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/audit"
	"github.com/mrPTqp/metrics/internal/contextkey"
)

type stubMetricWriter struct {
	gaugesSaved   map[string]float64
	countersSaved map[string]int64

	saveGaugeErr   error
	saveCounterErr error
	saveAllErr     error
}

func (s *stubMetricWriter) SaveGaugeMetric(_ context.Context, name string, value *float64) error {
	if s.gaugesSaved == nil {
		s.gaugesSaved = make(map[string]float64)
	}
	if value != nil {
		s.gaugesSaved[name] = *value
	}
	return s.saveGaugeErr
}

func (s *stubMetricWriter) SaveCounterMetric(_ context.Context, name string, value *int64) error {
	if s.countersSaved == nil {
		s.countersSaved = make(map[string]int64)
	}
	if value != nil {
		s.countersSaved[name] += *value
	}
	return s.saveCounterErr
}

func (s *stubMetricWriter) SaveAllMetrics(_ context.Context, gauges map[string]float64, counters map[string]int64) error {
	if s.gaugesSaved == nil {
		s.gaugesSaved = make(map[string]float64)
	}
	if s.countersSaved == nil {
		s.countersSaved = make(map[string]int64)
	}
	for k, v := range gauges {
		s.gaugesSaved[k] = v
	}
	for k, v := range counters {
		s.countersSaved[k] = v
	}
	return s.saveAllErr
}

type stubEventBus struct {
	published []audit.AuditEvent
	err       error
}

func (b *stubEventBus) Publish(e audit.AuditEvent) error {
	b.published = append(b.published, e)
	return b.err
}

func TestAuditService_SaveGaugeCounterAndAll(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := contextkey.WithLogger(context.Background(), logger)
	ctx = contextkey.WithClientIP(ctx, "127.0.0.1")

	writer := &stubMetricWriter{}

	svc := NewAuditService(writer, (*audit.EventBus)(nil))
	// Для проверки функциональности используем реальный eventBus
	svc.eventBus = (*audit.EventBus)(nil)

	// Проверяем, что методы writer вызываются
	gv := 1.23
	cv := int64(5)
	gauges := map[string]float64{"g1": gv}
	counters := map[string]int64{"c1": cv}

	if err := svc.SaveGaugeMetric(ctx, "g1", &gv); err != nil {
		t.Fatalf("SaveGaugeMetric() error = %v, want nil", err)
	}
	if writer.gaugesSaved["g1"] != gv {
		t.Fatalf("gauge saved = %v, want %v", writer.gaugesSaved["g1"], gv)
	}

	if err := svc.SaveCounterMetric(ctx, "c1", &cv); err != nil {
		t.Fatalf("SaveCounterMetric() error = %v, want nil", err)
	}
	if writer.countersSaved["c1"] != cv {
		t.Fatalf("counter saved = %v, want %v", writer.countersSaved["c1"], cv)
	}

	if err := svc.SaveAllMetrics(ctx, gauges, counters); err != nil {
		t.Fatalf("SaveAllMetrics() error = %v, want nil", err)
	}
	if len(writer.gaugesSaved) == 0 || len(writer.countersSaved) == 0 {
		t.Fatalf("expected SaveAllMetrics to save gauges and counters, got %#v %#v", writer.gaugesSaved, writer.countersSaved)
	}
}
