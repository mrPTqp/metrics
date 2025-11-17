package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/mrPTqp/metrics/internal/models"
)

type MetricsAgent struct {
	c  *http.Client
	ec *HTTPErrorClassifier
}

func NewMetricsAgent(client *http.Client) *MetricsAgent {
	return &MetricsAgent{
		c:  client,
		ec: NewHTTPErrorClassifier(),
	}
}

func (mh *MetricsAgent) StartMetricsAgent(address string, reportInterval, poolInterval int) {
	log.Printf("agent will send requests to %s", address)

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
			err := mh.sendMetrics(gauges, counters, mh.c, address)
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

func (mh *MetricsAgent) sendMetrics(gauges map[string]float64, counters map[string]int64, client *http.Client, address string) error {
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

	resp, err := mh.doWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("[ERROR] HTTP status " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func (mh *MetricsAgent) doWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	var maxRetries = 3
	var initialDelay = 1 * time.Second

	for i := 0; i <= maxRetries; i++ {
		resp, err = mh.c.Do(req)

		if err == nil {
			return resp, nil
		}

		classification := mh.ec.Classify(err)

		if classification == NonRetriable {
			return nil, err
		}

		if i == maxRetries {
			break
		}

		time.Sleep(initialDelay)
		initialDelay += 2
	}

	return nil, err
}
