package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/agent/repository"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/retry"
	"go.uber.org/zap"
)

// Агент-коллектор метрик
type MetricsAgent struct {
	c          *http.Client
	ec         *HTTPErrorClassifier
	cfg        *config.Config
	repository repository.MetricRepository
	logger     *zap.Logger
}

// Возвращает новый экземпляр MetricsAgent
func NewMetricsAgent(client *http.Client, cfg *config.Config, repository repository.MetricRepository, logger *zap.Logger) *MetricsAgent {
	return &MetricsAgent{
		c:          client,
		ec:         NewHTTPErrorClassifier(),
		cfg:        cfg,
		repository: repository,
		logger:     logger,
	}
}

// Собирает основные метрики
func (ma *MetricsAgent) PollMetrics() {
	_, counters := ma.repository.GetAllMetrics()
	newCounters := make(map[string]int64)
	newCounters["PollCount"] = counters["PollCount"] + 1
	newGauges := CollectGaugeMetrics()
	ma.repository.SaveAllMetrics(newGauges, newCounters)
}

// Собирает дополнительные метрики
func (ma *MetricsAgent) PollAdditionalGaugeMetrics() {
	newAdditionalGauges := CollectAdditionalGaugeMetrics()
	ma.repository.SaveAdditionalGaugeMetrics(newAdditionalGauges)
}

// Отправляет собранные метрики на сервер по HTTP
func (ma *MetricsAgent) SendMetrics(ctx context.Context) {
	address := ma.cfg.Address

	gauges, counters := ma.repository.GetAllMetrics()
	additionalGauges := ma.repository.GetAdditionalGaugeMetrics()

	if len(gauges) == 0 && len(counters) == 0 && len(additionalGauges) == 0 {
		ma.logger.Info("gauges and counters are empty")
		return
	}

	g := make([]models.Metrics, 0, len(gauges)+len(additionalGauges))
	for name, value := range gauges {
		req := models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		g = append(g, req)
	}
	for name, value := range additionalGauges {
		g = append(g, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		})
	}

	c := make([]models.Metrics, 0, len(counters))
	for name, delta := range counters {
		req := models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &delta,
		}
		c = append(c, req)
	}

	req := append(g, c...)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		ma.logger.Error("failed to marshal metrics", zap.Error(err))
		return
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		ma.logger.Error("failed to compress metrics", zap.Error(err))
		return
	}

	URL := "http://" + address.String() + "/updates"
	httpReq, err := http.NewRequest("POST", URL, bytes.NewReader(compressedBody))
	if err != nil {
		ma.logger.Error("failed to create HTTP request", zap.Error(err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Content-Encoding", "gzip")
	httpReq.Header.Set("Accept-Encoding", "gzip")

	var resp *http.Response
	err = retry.DoWithRetry(
		ctx,
		ma.ec,
		func() error {
			r, err2 := ma.c.Do(httpReq)
			if err2 != nil {
				return err2
			}
			defer r.Body.Close()

			resp = r
			return nil
		},
		3,
		1*time.Second,
	)

	if err != nil {
		ma.logger.Error("failed to send metrics after retries", zap.Error(err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		ma.logger.Error("HTTP request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("url", URL))
		return
	}

	ma.logger.Info("metrics successfully sent to server")
}

// Отправляет собранные метрики на сервер по gRPC
func (ma *MetricsAgent) SendMetricsGRPC(ctx context.Context, grpcAgent *GRPCMetricsAgent) {
	if grpcAgent == nil {
		ma.logger.Debug("gRPC agent is not configured")
		return
	}

	grpcAgent.SendMetrics(ctx, ma.repository)
}
