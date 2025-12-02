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
	
	a := agent.NewMetricsAgent(client, config, sugar)
	sugar.Infof("agent will start with params reportInterval: %d, poolInterval: %d", config.ReportInterval, config.PoolInterval)
	a.StartMetricsAgent()
}
