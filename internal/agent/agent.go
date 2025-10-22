package agent

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type MetricsAgent struct {
	client *http.Client
}

func NewMetricsAgent(client *http.Client) *MetricsAgent {
	return &MetricsAgent{
		client: &http.Client{},
	}
}

func (mh *MetricsAgent) StartMetricsAgent(address string, reportInterval, poolInterval int) {
	log.Printf("client will send requests to %s", address)

	var metrics map[string]float64
	var poolCounter = 0
	var lastReportTime = time.Now()
	for {
		metrics = collectMetrics()
		poolCounter++
		metrics["RandomValue"] = rand.Float64()

		currentTime := time.Now()
		if currentTime.Sub(lastReportTime) >= time.Duration(reportInterval)*time.Second {
			var errorCounter = 0
			err := sendMetrics(metrics, mh.client, poolCounter, address)
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

func sendMetrics(metrics map[string]float64, client *http.Client, poolCounter int, address string) error {
	for mName, mValue := range metrics {
		path := fmt.Sprintf("/update/%s/%s/%f", "gauge", mName, mValue)

		err := sendMetric(*client, "http://"+address+path)
		if err != nil {
			return err
		}
	}

	path := fmt.Sprintf("/update/%s/%s/%d", "counter", "PollCount", poolCounter)
	err := sendMetric(*client, "http://"+address+path)
	if err != nil {
		return err
	}
	return nil
}

func sendMetric(client http.Client, url string) error {
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("[ERROR] HTTP status " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

func collectMetrics() map[string]float64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := make(map[string]float64)
	metrics["GCCPUFraction"] = memStats.GCCPUFraction
	metrics["GCSys"] = float64(memStats.GCSys)
	metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	metrics["HeapIdle"] = float64(memStats.HeapIdle)
	metrics["HeapInuse"] = float64(memStats.HeapInuse)
	metrics["HeapObjects"] = float64(memStats.HeapObjects)
	metrics["HeapReleased"] = float64(memStats.HeapReleased)
	metrics["HeapSys"] = float64(memStats.HeapSys)
	metrics["LastGC"] = float64(memStats.LastGC)
	metrics["Lookups"] = float64(memStats.Lookups)
	metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	metrics["MCacheSys"] = float64(memStats.MCacheSys)
	metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	metrics["MSpanSys"] = float64(memStats.MSpanSys)
	metrics["Mallocs"] = float64(memStats.Mallocs)
	metrics["NextGC"] = float64(memStats.NextGC)
	metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	metrics["NumGC"] = float64(memStats.NumGC)
	metrics["OtherSys"] = float64(memStats.OtherSys)
	metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	metrics["StackInuse"] = float64(memStats.StackInuse)
	metrics["StackSys"] = float64(memStats.StackSys)
	metrics["Sys"] = float64(memStats.Sys)
	metrics["TotalAlloc"] = float64(memStats.TotalAlloc)
	return metrics
}
