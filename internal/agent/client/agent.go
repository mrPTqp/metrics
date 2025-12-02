package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/retry"
	"github.com/mrPTqp/metrics/internal/signer"

	"go.uber.org/zap"
)

type MetricsAgent struct {
	c      *http.Client
	ec     *HTTPErrorClassifier
	cfg    *config.Config
	logger *zap.SugaredLogger
}

// NewMetricsAgent принимает logger извне — не создаёт его самостоятельно
func NewMetricsAgent(client *http.Client, cfg *config.Config, logger *zap.SugaredLogger) *MetricsAgent {
	return &MetricsAgent{
		c:      client,
		ec:     NewHTTPErrorClassifier(),
		cfg:    cfg,
		logger: logger,
	}
}

func (mh *MetricsAgent) StartMetricsAgent() {
	address := mh.cfg.Address
	reportInterval := mh.cfg.ReportInterval
	poolInterval := mh.cfg.PoolInterval

	mh.logger.Infof("agent will send requests to %s", address.String())

	var poolCounter int64 = 0
	var lastReportTime = time.Now()
	for {
		gauges := CollectMetrics()
		gauges["RandomValue"] = rand.Float64()
		counters := make(map[string]int64)
		poolCounter++
		counters["PollCount"] = poolCounter

		currentTime := time.Now()
		if currentTime.Sub(lastReportTime) >= time.Duration(reportInterval)*time.Second {
			err := mh.sendMetrics(gauges, counters, address.String())
			if err != nil {
				mh.logger.Errorf("failed to send metrics: %v", err)
			} else {
				mh.logger.Infof("metrics successfully sent")
			}
			lastReportTime = currentTime
		}

		time.Sleep(time.Duration(poolInterval) * time.Second)
	}
}

func (mh *MetricsAgent) sendMetrics(gauges map[string]float64, counters map[string]int64, address string) error {
	if len(gauges) == 0 && len(counters) == 0 {
		return nil
	}

	g := make([]models.Metrics, 0, len(gauges))
	for name, value := range gauges {
		g = append(g, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		})
	}

	c := make([]models.Metrics, 0, len(counters))
	for name, delta := range counters {
		c = append(c, models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &delta,
		})
	}

	req := append(g, c...)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		mh.logger.Errorf("failed to marshal metrics: %v", err)
		return err
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		mh.logger.Errorf("failed to compress metrics: %v", err)
		return err
	}

	var signature *string
	if mh.cfg.SecretKey != nil && *mh.cfg.SecretKey != "" {
		signature, err = signer.Sign(compressedBody, mh.cfg.SecretKey)
		if err != nil {
			mh.logger.Errorf("failed to sign request: %v", err)
			return err
		}
	}

	url := "http://" + address + "/updates"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressedBody))
	if err != nil {
		mh.logger.Errorf("failed to create HTTP request: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Content-Encoding", "gzip")
	if signature != nil {
		httpReq.Header.Set("HashSHA256", *signature)
	}

	var resp *http.Response
	err = retry.DoWithRetry(
		context.Background(),
		mh.ec,
		func() error {
			r, err := mh.c.Do(httpReq)
			if err != nil {
				return err
			}
			defer r.Body.Close()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}

			// Проверка подписи ответа — используем mh.logger
			respSignature := r.Header.Get("HashSHA256")
			if respSignature != "" {
				if !signer.Verify(body, &respSignature, mh.cfg.SecretKey, mh.logger) {
					mh.logger.Errorf("response signature verification failed")
					return errors.New("response signature verification failed")
				}
			}

			resp = &http.Response{
				StatusCode: r.StatusCode,
				Header:     r.Header,
				Body:       io.NopCloser(bytes.NewReader(body)),
			}
			return nil
		},
		3,
		1*time.Second,
	)
	if err != nil {
		mh.logger.Errorf("request failed after retries: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		mh.logger.Errorf("HTTP request failed with status: %d", resp.StatusCode)
		return errors.New("HTTP status " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
