package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mrPTqp/metrics/internal/agent"
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

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	a := agent.NewMetricsAgent(client)
	log.Printf("agent will start with params reportInterval: %d, poolInterval: %d", reportInterval, poolInterval)
	a.StartMetricsAgent(address.String(), reportInterval, poolInterval)
}
