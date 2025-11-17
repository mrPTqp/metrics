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

func (ps *PostgresStorage) withRetry(ctx context.Context, operation func() error) error {
	const maxRetries = 3
	delay := 1 * time.Second

	var err error
	for i := 0; i <= maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if ps.classifier.Classify(err) == NonRetriable {
			return err
		}

		if i == maxRetries {
			return err
		}

		time.Sleep(delay)
		delay += 2
	}

	return err
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
	return ps.withRetry(ctx, func() error {
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
	return ps.withRetry(ctx, func() error {
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
	err := ps.withRetry(ctx, func() error {
		return ps.DB.QueryRowContext(ctx,
			`SELECT value 
			FROM gauges 
			WHERE name = $1`,
			name).Scan(&value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || !value.Valid {
			return 0.0, ErrMetricNotFound
		}
		return 0.0, err
	}
	if !value.Valid {
		return 0.0, ErrMetricNotFound
	}
	return value.Float64, nil
}

func (ps *PostgresStorage) GetCounter(name string) (int64, error) {
	ctx := context.Background()
	var value sql.NullInt64
	err := ps.withRetry(ctx, func() error {
		return ps.DB.QueryRowContext(ctx,
			`SELECT value 
			FROM counters 
			WHERE name = $1`,
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
	return ps.withRetry(ctx, func() error {
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
	err := ps.withRetry(context.Background(), ps.CheckConnection)
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
