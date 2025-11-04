package main

import (
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
	"go.uber.org/zap"
)

type NetAddress struct {
	Host string
	Port int
}

var address NetAddress = NetAddress{"localhost", 8080}
var reportInterval int = 10
var poolInterval int = 2

func main() {
	parseFlags()
	parseEnvs()

	var sugar *zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = logger.Sugar()

	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	a := agent.NewMetricsAgent(client)
	sugar.Infof("agent will start with params reportInterval: %d, poolInterval: %d", reportInterval, poolInterval)
	a.StartMetricsAgent(address.String(), reportInterval, poolInterval)
}
