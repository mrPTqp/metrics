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

	client.Transport = &agent.ExtraHeadersTransport{
		RoundTripper: client.Transport,
		Logger:       log,
	}

	mr := storage.NewMemStorage()
	a := agent.NewMetricsAgent(client, cfg, mr, log)

	var grpcAgent *agent.GRPCMetricsAgent
	if cfg.GRPCEnabled {
		grpcMetadataTransport := agent.NewGRPCMetadataTransport(log)
		
		var err error
		grpcAgent, err = agent.NewGRPCMetricsAgent(cfg, log, grpcMetadataTransport)
		if err != nil {
			log.Fatal("failed to create gRPC agent", zap.Error(err))
		} else {
			log.Info("gRPC agent created successfully")
		}
	} else {
		log.Info("gRPC agent disabled")
	}

	sc := scheduler.NewScheduler(a, grpcAgent, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	sendWorkersSize := cfg.RateLimit
	taskChannelSize := sendWorkersSize * 3
	go func() {
		sc.Start(ctx, cfg.PollInterval, cfg.ReportInterval, sendWorkersSize, taskChannelSize)
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")
	log.Info("sending final metrics...")
	a.SendMetrics(ctx)
	log.Info("waiting for scheduler to finish...")
	log.Info("agent stopped gracefully")
}
