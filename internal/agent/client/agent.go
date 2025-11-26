package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/retry"
)

type MetricsAgent struct {
	c   *http.Client
	ec  *HTTPErrorClassifier
	cfg *config.Config
}

func NewMetricsAgent(client *http.Client, cfg *config.Config) *MetricsAgent {
	return &MetricsAgent{
		c:   client,
		ec:  NewHTTPErrorClassifier(),
		cfg: cfg,
	}
}

func (mh *MetricsAgent) StartMetricsAgent() {
	address := mh.cfg.Address
	reportInterval := mh.cfg.ReportInterval
	poolInterval := mh.cfg.PoolInterval

	log.Printf("agent will send requests to %s", address.String())

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
			var errorCounter = 0
			err := mh.sendMetrics(gauges, counters, address.String())
			if err != nil {
				errorCounter++
				log.Printf("[ERROR] %s", err)
			}
			log.Printf("all metrics sent. errors number %d", errorCounter)
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
		req := models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		g = append(g, req)
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
		return err
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		return err
	}

	url := "http://" + address + "/updates"

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressedBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Content-Encoding", "gzip")	

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
			resp = r
			return nil
		},
		3,
		1*time.Second,
	)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("[ERROR] HTTP status " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
