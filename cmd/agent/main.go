package main

import (
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
	"github.com/mrPTqp/metrics/internal/logger"
)

func main() {
	sugar := logger.NewSugarLogger()

	config := LoadConfig()

	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	a := agent.NewMetricsAgent(client)
	sugar.Infof("agent will start with params reportInterval: %d, poolInterval: %d", config.ReportInterval, config.PoolInterval)
	a.StartMetricsAgent(config.Address.String(), config.ReportInterval, config.PoolInterval)
}
