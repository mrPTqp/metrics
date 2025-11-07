package main

import (
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
	"go.uber.org/zap"
)



func main() {
	var sugar *zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()

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
