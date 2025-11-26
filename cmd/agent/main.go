package main

import (
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/client"
	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/logger"
)

func main() {
	sugar := logger.NewSugarLogger()

	config := config.LoadConfig()

	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	if config.SecretKey != nil && *config.SecretKey != "" {
		originalTransport := client.Transport
		if originalTransport == nil {
			originalTransport = http.DefaultTransport
		}

		client.Transport = &agent.SigningTransport{
			RoundTripper: originalTransport,
			SecretKey: *config.SecretKey,
			Logger: sugar,
		}
	}

	a := agent.NewMetricsAgent(client, config)
	sugar.Infof("agent will start with params reportInterval: %d, poolInterval: %d", config.ReportInterval, config.PoolInterval)
	a.StartMetricsAgent()
}
