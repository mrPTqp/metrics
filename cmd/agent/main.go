package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrPTqp/metrics/internal/agent/client"
	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/agent/scheduler"
	"github.com/mrPTqp/metrics/internal/agent/storage"
	"github.com/mrPTqp/metrics/internal/logger"
)

func main() {
	sugar := logger.NewSugarLogger()

	config := config.LoadConfig()
	sugar.Infow("configuration created", "config", config)

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
			SecretKey:    *config.SecretKey,
			Logger:       sugar,
		}
	}

	mr := storage.NewMemStorage()
	a := agent.NewMetricsAgent(client, config, mr, sugar)
	sc := scheduler.NewScheduler(a, sugar)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sendWorkersSize := config.RateLimit
	taskChannelSize := sendWorkersSize * 3
	go func() {
		sc.Start(ctx, config.PollInterval, config.ReportInterval, sendWorkersSize, taskChannelSize)
	}()

	<-ctx.Done()
	sugar.Info("Shutdown signal received")
	sugar.Info("Waiting for scheduler to finish...")
	sugar.Info("Agent stopped gracefully")
}
