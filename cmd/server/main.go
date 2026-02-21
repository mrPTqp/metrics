package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrPTqp/metrics/internal/logger"
	"github.com/mrPTqp/metrics/internal/server/app"
	"github.com/mrPTqp/metrics/internal/server/bootstrap"
	"github.com/mrPTqp/metrics/internal/server/config"
	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger()
	defer func() {
		_ = log.Sync()
	}()

	cfg := config.LoadConfig()
	log.Info("configuration created", zap.Any("config", cfg))

	bootstrapper := bootstrap.NewBootstrapper(cfg, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var components *bootstrap.AppComponents
	var err error
	if components, err = bootstrapper.MustRun(ctx); err != nil {
		log.Fatal("failed to bootstrap application", zap.Error(err))
	}

	application := app.NewApp(components)

	go application.RunWithContext(ctx)

	log.Info("application started")

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	application.Shutdown(shutdownCtx)
	log.Info("application stopped")
}
