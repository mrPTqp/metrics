package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	agent "github.com/mrPTqp/metrics/internal/agent/client"
	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/agent/scheduler"
	"github.com/mrPTqp/metrics/internal/agent/storage"
	"github.com/mrPTqp/metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger()
	defer func() {
		_ = log.Sync()
	}()

	cfg := config.LoadConfig()
	log.Info("configuration created", zap.Any("config", cfg))

	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	if cfg.SecretKey != nil && *cfg.SecretKey != "" {
		originalTransport := client.Transport
		if originalTransport == nil {
			originalTransport = http.DefaultTransport
		}

		client.Transport = &agent.SigningTransport{
			RoundTripper: originalTransport,
			SecretKey:    *cfg.SecretKey,
			Logger:       log,
		}
	}

	mr := storage.NewMemStorage()
	a := agent.NewMetricsAgent(client, cfg, mr, log)
	sc := scheduler.NewScheduler(a, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sendWorkersSize := cfg.RateLimit
	taskChannelSize := sendWorkersSize * 3
	go func() {
		sc.Start(ctx, cfg.PollInterval, cfg.ReportInterval, sendWorkersSize, taskChannelSize)
	}()

	<-ctx.Done()
	log.Info("Shutdown signal received")
	log.Info("Waiting for scheduler to finish...")
	log.Info("Agent stopped gracefully")
}
