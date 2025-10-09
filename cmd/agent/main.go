package main

import (
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
)

func main() {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	a := agent.NewMetricsAgent(client)
	a.StartMetricsAgent()
}