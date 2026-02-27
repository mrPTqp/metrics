package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	agent "github.com/mrPTqp/metrics/internal/agent/client"
	"github.com/mrPTqp/metrics/internal/agent/config"
	"github.com/mrPTqp/metrics/internal/agent/scheduler"
	"github.com/mrPTqp/metrics/internal/agent/storage"
	"github.com/mrPTqp/metrics/internal/crypto"
	"github.com/mrPTqp/metrics/internal/logger"
	"go.uber.org/zap"
)

var (
	buildVersion, buildDate, buildCommit string
)

func printBuildInfo(w io.Writer) {
	v := func(val string) string {
		if val != "" {
			return val
		}
		return "N/A"
	}
	fmt.Fprintf(w, "Build version: %s\n", v(buildVersion))
	fmt.Fprintf(w, "Build date: %s\n", v(buildDate))
	fmt.Fprintf(w, "Build commit: %s\n", v(buildCommit))
}

func main() {
	printBuildInfo(os.Stdout)

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
		client.Transport = &agent.SigningTransport{
			RoundTripper: client.Transport,
			SecretKey:    *cfg.SecretKey,
			Logger:       log,
		}
	}

	if cfg.CertPath != nil && *cfg.CertPath != "" {
		cert, err := crypto.ReadCertificate(*cfg.CertPath)
		if err != nil || cert == nil {
			log.Fatal("failed to load certificate", zap.Error(err))
		}

		client.Transport = &agent.EncryptTransport{
			RoundTripper: client.Transport,
			Cert:         cert,
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
	log.Info("shutdown signal received")
	log.Info("waiting for scheduler to finish...")
	log.Info("agent stopped gracefully")
}
