package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
)

const pollInterval = 2
const reportInterval = 10

func main() {
	client := http.Client{
		Timeout: time.Second * 30,
	}

	var metrics map[string]float64
	var poolCounter int = 0
	var lastReportTime = time.Now()
	for {
		metrics = agent.CollectMetrics()
		poolCounter++
		metrics["RandomValue"] = rand.Float64()

		currentTime := time.Now()
		if currentTime.Sub(lastReportTime) >= reportInterval*time.Second {
			var errorCounter int = 0
			err := sendMetrics(metrics, client, poolCounter)
			if err != nil {
				errorCounter++
				log.Printf("[ERROR] %s", err)
			}
			log.Printf("all metrics sent. errors number %d", errorCounter)
			lastReportTime = currentTime
		}

		time.Sleep(pollInterval * time.Second)
	}
}

func sendMetrics(metrics map[string]float64, client http.Client, poolCounter int) error {
	for mName, mValue := range metrics {
		path := fmt.Sprintf("/update/%s/%s/%f", "gauge", mName, mValue)

		err := sendMetric(client, path)
		if err != nil {
			return err
		}
	}

	path := fmt.Sprintf("/update/%s/%s/%d", "counter", "PollCount", poolCounter)
	err := sendMetric(client, path)
	if err != nil {
		return err
	}
	return nil
}

func sendMetric(client http.Client, path string) error {
	resp, err := client.Post("http://localhost:8080"+path, "text/plain", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("[ERROR] HTTP status " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}


