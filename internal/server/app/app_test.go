package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/mrPTqp/metrics/internal/audit"
	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/server/backup"
	"github.com/mrPTqp/metrics/internal/server/bootstrap"
	"github.com/mrPTqp/metrics/internal/server/config"
	"github.com/mrPTqp/metrics/internal/server/repository"
	"github.com/mrPTqp/metrics/internal/server/storage"
)

// stubMetricRepository is a simple in-memory implementation of MetricRepository
// used only inside tests to track Close calls.
type stubMetricRepository struct {
	repository.MetricRepository
	closeCalls int32
}

func (s *stubMetricRepository) Close() error {
	atomic.AddInt32(&s.closeCalls, 1)
	return nil
}

// shutdownAwareProcessor implements both AuditProcessor and Shutdown to be used in App.Shutdown tests.
type shutdownAwareProcessor struct {
	calls int32
}

func (p *shutdownAwareProcessor) Write(audit.AuditEvent) error {
	return nil
}

func (p *shutdownAwareProcessor) Shutdown(context.Context) error {
	atomic.AddInt32(&p.calls, 1)
	return nil
}

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	var addr models.NetAddress
	if err := addr.SetAddress("127.0.0.1:0"); err != nil {
		t.Fatalf("SetAddress() error = %v", err)
	}

	return &config.Config{
		Address:       addr,
		StoreInterval: 1,
		BackupFilePath: t.TempDir() + "/backup.json",
	}
}

func newTestAppComponents(t *testing.T) *bootstrap.AppComponents {
	t.Helper()

	logger := zaptest.NewLogger(t)
	cfg := newTestConfig(t)

	memRepo := storage.NewMemStorage()

	return &bootstrap.AppComponents{
		Config:  cfg,
		Logger:  logger,
		Repo:    memRepo,
		Handler: nil,
	}
}

func TestWrap_TableDriven(t *testing.T) {
	type mwCase struct {
		name        string
		middlewares []func(http.Handler) http.Handler
	}

	var calledOrder []string

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledOrder = append(calledOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	logMiddleware := func(id string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calledOrder = append(calledOrder, id)
				next.ServeHTTP(w, r)
			})
		}
	}

	tests := []mwCase{
		{
			name:        "no middlewares",
			middlewares: nil,
		},
		{
			name:        "single middleware",
			middlewares: []func(http.Handler) http.Handler{logMiddleware("mw1")},
		},
		{
			name:        "multiple middlewares",
			middlewares: []func(http.Handler) http.Handler{logMiddleware("mw1"), logMiddleware("mw2")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calledOrder = calledOrder[:0]

			wrapped := wrap(baseHandler, tt.middlewares...)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			if len(calledOrder) == 0 || calledOrder[len(calledOrder)-1] != "handler" {
				t.Fatalf("last call should be handler, got sequence %#v", calledOrder)
			}
		})
	}
}

func TestApp_RunBackgroundJobs_StopsOnContextDone(t *testing.T) {
	components := newTestAppComponents(t)
	app := NewApp(components)

	// Speed up ticker for test and make sure background loop is entered.
	app.ticker = time.NewTicker(10 * time.Millisecond)
	defer app.ticker.Stop()
	app.backuper = new(backup.Backuper)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		app.runBackgroundJobs(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("runBackgroundJobs did not return after context cancellation")
	}
}

func TestApp_Shutdown_IsIdempotentAndClosesResources(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := newTestConfig(t)

	repo := &stubMetricRepository{}

	proc := &shutdownAwareProcessor{}
	auditProcessors := []audit.AuditProcessor{proc}

	app := &App{
		cfg: &bootstrap.AppComponents{
			Config:          cfg,
			Logger:          logger,
			Repo:            repo,
			EventBus:        nil,
			AuditProcessors: auditProcessors,
		},
		server: &http.Server{
			Addr: cfg.Address.String(),
		},
		ticker: time.NewTicker(time.Millisecond),
		repo:   repo,
	}
	defer app.ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	app.Shutdown(ctx)
	app.Shutdown(ctx)

	if got := atomic.LoadInt32(&repo.closeCalls); got != 1 {
		t.Fatalf("repo.Close() calls = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&proc.calls); got != 1 {
		t.Fatalf("audit processor Shutdown() calls = %d, want 1", got)
	}
}

