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

type MetricsAgent struct {
	c          *http.Client
	ec         *HTTPErrorClassifier
	cfg        *config.Config
	repository repository.MetricRepository
	logger     *zap.SugaredLogger
}

func NewMetricsAgent(client *http.Client, cfg *config.Config, repository repository.MetricRepository, logger *zap.SugaredLogger) *MetricsAgent {
	return &MetricsAgent{
		c:          client,
		ec:         NewHTTPErrorClassifier(),
		cfg:        cfg,
		repository: repository,
		logger:     logger,
	}
}

func (ma *MetricsAgent) PollMetrics() {
	_, counters := ma.repository.GetAllMetrics()
	newCounters := make(map[string]int64)
	newCounters["PollCount"] = counters["PollCount"] + 1
	newGauges := CollectGaugeMetrics()
	ma.repository.SaveAllMetrics(newGauges, newCounters)
}

func (ma *MetricsAgent) PollAdditionalGaugeMetrics() {
	newAdditionalGauges := CollectAdditionalGaugeMetrics()
	ma.repository.SaveAdditionalGaugeMetrics(newAdditionalGauges)
}

func (ma *MetricsAgent) SendMetrics() {
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
		ma.logger.Errorf("failed to marshal metrics: %v", err)
		return
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		return
	}

	url := "http://" + address.String() + "/updates"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressedBody))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Content-Encoding", "gzip")
	httpReq.Header.Set("Accept-Encoding", "gzip")

	var resp *http.Response
	err = retry.DoWithRetry(
		context.Background(),
		ma.ec,
		func() error {
			r, err := ma.c.Do(httpReq)
			if err != nil {
				return err
			}
			defer r.Body.Close()

			resp = r
			return nil
		},
		3,
		1*time.Second,
	)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		ma.logger.Errorf("HTTP request failed with status: %d", resp.StatusCode)
		return
	}

	ma.logger.Info("Metrics successfully sent to server")
}
