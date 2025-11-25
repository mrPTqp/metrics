// internal/server/storage/postgres_storage.go

package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/retry"
)

var ErrMetricNotFound = errors.New("metric not found")

type PostgresStorage struct {
	DB         *sql.DB
	logger     *zap.SugaredLogger
	classifier *PostgresErrorClassifier
}

func NewPostgresStorage(databaseDsn string, logger *zap.SugaredLogger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		DB:         db,
		logger:     logger,
		classifier: NewPostgresErrorClassifier(),
	}, nil
}

func (ps *PostgresStorage) doWithRetry(ctx context.Context, operation func() error) error {
	return retry.DoWithRetry(
		ctx,
		ps.classifier,
		operation,
		3,
		1*time.Second,
	)
}

func extractPGCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return "N/A"
}

func (ps *PostgresStorage) SaveGauge(name string, value *float64) error {
	ctx := context.Background()
	return ps.doWithRetry(ctx, func() error {
		_, err := ps.DB.ExecContext(ctx, `
			INSERT INTO gauges (name, value) 
			VALUES ($1, $2) 
			ON CONFLICT (name) 
			DO UPDATE SET value = EXCLUDED.value`,
			name, value)
		return err
	})
}

func (ps *PostgresStorage) SaveCounter(name string, value *int64) error {
	ctx := context.Background()
	return ps.doWithRetry(ctx, func() error {
		_, err := ps.DB.ExecContext(ctx, `
			INSERT INTO counters (name, value) 
			VALUES ($1, $2) 
			ON CONFLICT (name) 
			DO UPDATE SET value = counters.value + EXCLUDED.value`,
			name, value)
		return err
	})
}

func (ps *PostgresStorage) GetGauge(name string) (float64, error) {
	ctx := context.Background()
	var value sql.NullFloat64
	err := ps.doWithRetry(ctx, func() error {
		return ps.DB.QueryRowContext(ctx,
			`SELECT value FROM gauges WHERE name = $1`,
			name).Scan(&value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || !value.Valid {
			return 0, ErrMetricNotFound
		}
		return 0, err
	}
	if !value.Valid {
		return 0, ErrMetricNotFound
	}
	return value.Float64, nil
}

func (ps *PostgresStorage) GetCounter(name string) (int64, error) {
	ctx := context.Background()
	var value sql.NullInt64
	err := ps.doWithRetry(ctx, func() error {
		return ps.DB.QueryRowContext(ctx,
			`SELECT value FROM counters WHERE name = $1`,
			name).Scan(&value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || !value.Valid {
			return 0, ErrMetricNotFound
		}
		return 0, err
	}
	if !value.Valid {
		return 0, ErrMetricNotFound
	}
	return value.Int64, nil
}

func (ps *PostgresStorage) ListGauges() (map[string]float64, error) {
	ctx := context.Background()
	rows, err := ps.DB.QueryContext(ctx, "SELECT name, value FROM gauges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value sql.NullFloat64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		if value.Valid {
			result[name] = value.Float64
		}
	}
	return result, rows.Err()
}

func (ps *PostgresStorage) ListCounters() (map[string]int64, error) {
	ctx := context.Background()
	rows, err := ps.DB.QueryContext(ctx, "SELECT name, value FROM counters")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var name string
		var value sql.NullInt64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		if value.Valid {
			result[name] = value.Int64
		}
	}
	return result, rows.Err()
}

func (ps *PostgresStorage) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	ctx := context.Background()
	return ps.doWithRetry(ctx, func() error {
		tx, err := ps.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		for name, value := range gauges {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO gauges (name, value) 
				VALUES ($1, $2) 
				ON CONFLICT (name) 
				DO UPDATE SET value = EXCLUDED.value`,
				name, value)
			if err != nil {
				return err
			}
		}

		for name, value := range counters {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO counters (name, value) 
				VALUES ($1, $2) 
				ON CONFLICT (name) 
				DO UPDATE SET value = counters.value + EXCLUDED.value`,
				name, value)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}

func (ps *PostgresStorage) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return ps.DB.PingContext(ctx)
}

func (ps *PostgresStorage) CheckStorageAvailability() bool {
	err := ps.doWithRetry(context.Background(), ps.CheckConnection)
	return err == nil
}

func (ps *PostgresStorage) Close() error {
	if ps.DB != nil {
		err := ps.DB.Close()
		if err != nil {
			return err
		}
	}
	ps.logger.Info("Postgres storage closed")
	return nil
}
