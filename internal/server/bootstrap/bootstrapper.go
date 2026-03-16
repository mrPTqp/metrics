package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mrPTqp/metrics/internal/audit"
	"github.com/mrPTqp/metrics/internal/crypto"
	"github.com/mrPTqp/metrics/internal/proto"
	grpcInterceptor "github.com/mrPTqp/metrics/internal/server/api/grpc"
	grpcHandler "github.com/mrPTqp/metrics/internal/server/api/grpc/handler"
	httpHandler "github.com/mrPTqp/metrics/internal/server/api/http/handler"
	"github.com/mrPTqp/metrics/internal/server/api/http/middleware"
	"github.com/mrPTqp/metrics/internal/server/backup"
	"github.com/mrPTqp/metrics/internal/server/config"
	"github.com/mrPTqp/metrics/internal/server/repository"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"github.com/mrPTqp/metrics/internal/server/storage/migrations"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Выделенный слой быстрой инициализации компонентов для запуска
type Bootstrapper struct {
	cfg    *config.Config
	logger *zap.Logger
}

// Возвращает новый экземпляр Bootstrapper
func NewBootstrapper(cfg *config.Config, logger *zap.Logger) *Bootstrapper {
	return &Bootstrapper{cfg: cfg, logger: logger}
}

// Возвращает компоненты для запуска
func (bs *Bootstrapper) MustRun(ctx context.Context) (*AppComponents, error) {
	bs.logger.Info("starting application bootstrap...")

	var mr repository.MetricRepository
	if bs.cfg.DatabaseDsn != nil && *bs.cfg.DatabaseDsn != "" {
		bs.logger.Info("applying database migrations...")
		if err := migrations.RunMigrations(*bs.cfg.DatabaseDsn, bs.logger); err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
		bs.logger.Info("migrations applied successfully or no changes")

		var err error
		mr, err = storage.NewPostgresStorage(*bs.cfg.DatabaseDsn, bs.logger)
		if err != nil {
			return nil, fmt.Errorf("init postgres storage error: %w", err)
		}

		checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		ok := mr.CheckStorageAvailability(checkCtx)
		cancel()
		if !ok {
			return nil, fmt.Errorf("postgres connection check failed: %w", err)
		}
	} else {
		mr = storage.NewMemStorage()
	}

	baseService := service.NewMetricsService(mr, bs.logger)
	var writer service.MetricWriter = baseService
	var reader service.MetricReader = baseService
	var lister service.MetricsLister = baseService
	var pinger service.Pinger = baseService

	p := storage.NewFileProducer(bs.cfg.BackupFilePath, bs.logger)
	c := storage.NewFileConsumer(bs.cfg.BackupFilePath, bs.logger)
	fs := storage.NewFileStorage(p, c, mr, bs.cfg.SyncBackupToFile, bs.logger)

	var b *backup.Backuper
	if bs.cfg.SyncBackupToFile {
		writer = service.NewFileBackupService(baseService, fs, bs.logger)
	} else {
		b = backup.NewBackuper(writer, lister, fs, bs.logger)
	}

	var auditProcessors []audit.AuditProcessor
	var eventBus *audit.EventBus
	if bs.cfg.FileAuditEnabled {
		proc, err := audit.NewFileAuditProcessor(bs.cfg.AuditFilePath, bs.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create file audit processor: %w", err)
		}
		auditProcessors = append(auditProcessors, proc)
		bs.logger.Info("file audit enabled", zap.String("path", bs.cfg.AuditFilePath))
	}
	if bs.cfg.HTTPAuditEnabled {
		proc := audit.NewHTTPAuditProcessor(*bs.cfg.AuditURL, bs.logger)
		auditProcessors = append(auditProcessors, proc)
		bs.logger.Info("HTTP audit enabled", zap.String("url", *bs.cfg.AuditURL))
	}
	if len(auditProcessors) > 0 {
		eventBus = audit.NewEventBus(auditProcessors, 1000, bs.logger)
		writer = service.NewAuditService(writer, eventBus)
	}

	if bs.cfg.Restore {
		r := backup.NewRestorer(writer, fs, bs.logger)
		restoreCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := r.Restore(restoreCtx)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to restore metrics: %w", err)
		}
	}

	mh := httpHandler.NewMetricHandler(writer, reader, lister, pinger, bs.logger)

	mws := []func(http.Handler) http.Handler{
		middleware.LoggingRequestBodyMiddleware,
		middleware.GzipMiddleware,
	}

	if bs.cfg.SecretKey != nil && *bs.cfg.SecretKey != "" {
		mws = append(mws, middleware.SignMiddleware(*bs.cfg.SecretKey))
	}

	if bs.cfg.PrivateKeyPath != nil && *bs.cfg.PrivateKeyPath != "" {
		privateKey, err := crypto.ReadPrivateKey(*bs.cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		mws = append(mws, middleware.DecryptMiddleware(privateKey))
	}

	mws = append(mws, middleware.LoggingMiddleware(bs.logger))

	if bs.cfg.TrustedSubnet != nil {
		mws = append(mws, middleware.SubnetMiddleware(bs.cfg.TrustedSubnet, bs.logger))
	}

	var grpcServer *grpc.Server
	if bs.cfg.GRPCEnabled {
		grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpcInterceptor.SubnetInterceptor(bs.cfg.TrustedSubnet, bs.logger)),
		)
		
		proto.RegisterMetricsServer(grpcServer, grpcHandler.NewMetricsHandler(writer, bs.logger))
		reflection.Register(grpcServer)
	}

	return &AppComponents{
		Config:          bs.cfg,
		Logger:          bs.logger,
		Repo:            mr,
		Handler:         mh,
		Writer:          writer,
		Backuper:        b,
		Middlewares:     mws,
		EventBus:        eventBus,
		AuditProcessors: auditProcessors,
		GRPCServer:      grpcServer,
	}, nil
}

type AppComponents struct {
	Config          *config.Config
	Logger          *zap.Logger
	Repo            repository.MetricRepository
	Handler         *httpHandler.MetricHandler
	Writer          service.MetricWriter
	Backuper        *backup.Backuper
	Middlewares     []func(http.Handler) http.Handler
	EventBus        *audit.EventBus
	AuditProcessors []audit.AuditProcessor
	GRPCServer      *grpc.Server
}
