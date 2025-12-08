package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
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

	var signature string
	if mh.cfg.SecretKey != nil && *mh.cfg.SecretKey != "" {
		signature, err = signer.Sign(jsonBody, *mh.cfg.SecretKey)
		if err != nil {
			mh.logger.Errorf("failed to sign request: %v", err)
			return err
		}
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		mh.logger.Errorf("failed to compress metrics: %v", err)
		return err
	}

	url := "http://" + address + "/updates"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressedBody))
	if err != nil {
		mh.logger.Errorf("failed to create HTTP request: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Content-Encoding", "gzip")
	if signature != "" {
		httpReq.Header.Set("HashSHA256", signature)
	}

	mh.logRequest(httpReq, jsonBody)

	var resp *http.Response
	err = retry.DoWithRetry(
		context.Background(),
		mh.ec,
		func() error {
			r, err := mh.c.Do(httpReq)
			if err != nil {
				return err
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				r.Body.Close()
				return err
			}
			r.Body.Close()

			var decompressed []byte
			contentEncoding := r.Header.Get("Content-Encoding")
			if strings.Contains(contentEncoding, "gzip") {
				decompressed, err = Decompress(body)
				if err != nil {
					return err
				}
			} else {
				decompressed = body
			}

			respSignature := r.Header.Get("HashSHA256")
			if respSignature != "" {
				if !signer.Verify(decompressed, respSignature, *mh.cfg.SecretKey, mh.logger) {
					mh.logger.Error("response signature verification failed")
					return errors.New("response signature verification failed")
				}
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			resp = r
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

func (mh *MetricsAgent) logRequest(req *http.Request, originalBody []byte) {
	bodyStr := string(originalBody)
	mh.logger.Infow("outgoing request",
		"method", req.Method,
		"url", req.URL.String(),
		"headers", req.Header,
		"body", bodyStr,
	)
}

func (mh *MetricsAgent) logResponse(resp *http.Response, compressedBody []byte) {
	var bodyStr string
	contentEncoding := resp.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		decompressed, err := Decompress(compressedBody)
		if err != nil {
			bodyStr = fmt.Sprintf("<failed to decompress: %v>", err)
		} else {
			bodyStr = string(decompressed)
		}
	} else {
		bodyStr = string(compressedBody)
	}

	mh.logger.Infow("incoming response",
		"status", resp.Status,
		"statusCode", resp.StatusCode,
		"headers", resp.Header,
		"body", bodyStr,
	)
}
