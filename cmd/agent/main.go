package main

import (
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
)

func main() {
	parseFlags()

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	a := agent.NewMetricsAgent(client)
	a.StartMetricsAgent(address.String(), reportInterval, poolInterval)
}
