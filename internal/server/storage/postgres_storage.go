package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/retry"
)

var ErrMetricNotFound = errors.New("metric not found")

type PostgresStorage struct {
	DB         *sql.DB
	logger     *zap.Logger
	classifier *PostgresErrorClassifier
}

func NewPostgresStorage(databaseDsn string, logger *zap.Logger) (*PostgresStorage, error) {
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

func (ps *PostgresStorage) doWithRetry(ctx context.Context, operation func(context.Context) error) error {
	log := contextkey.LoggerFromContext(ctx)
	return retry.DoWithRetry(
		ctx,
		ps.classifier,
		func() error {
			err := operation(ctx)
			if err != nil {
				log.Warn("operation failed, retrying...", zap.Error(err))
			}
			return err
		},
		3,
		1*time.Second,
	)
}

func (ps *PostgresStorage) CheckConnection(ctx context.Context) error {
	return ps.DB.PingContext(ctx)
}

func (ps *PostgresStorage) SaveGauge(ctx context.Context, name string, value *float64) error {
	log := contextkey.LoggerFromContext(ctx)
	err := ps.doWithRetry(ctx, func(ctx context.Context) error {
		_, err := ps.DB.ExecContext(ctx, `
			INSERT INTO gauges (name, value) 
			VALUES ($1, $2) 
			ON CONFLICT (name) 
			DO UPDATE SET value = EXCLUDED.value`,
			name, value)
		return err
	})
	if err != nil {
		log.Error("failed to save gauge", zap.String("name", name), zap.Error(err))
	}
	return err
}

func (ps *PostgresStorage) SaveCounter(ctx context.Context, name string, value *int64) error {
	log := contextkey.LoggerFromContext(ctx)
	err := ps.doWithRetry(ctx, func(ctx context.Context) error {
		_, err := ps.DB.ExecContext(ctx, `
			INSERT INTO counters (name, value) 
			VALUES ($1, $2) 
			ON CONFLICT (name) 
			DO UPDATE SET value = counters.value + EXCLUDED.value`,
			name, value)
		return err
	})
	if err != nil {
		log.Error("failed to save counter", zap.String("name", name), zap.Error(err))
	}
	return err
}

func (ps *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	log := contextkey.LoggerFromContext(ctx)
	var value sql.NullFloat64
	err := ps.doWithRetry(ctx, func(ctx context.Context) error {
		return ps.DB.QueryRowContext(ctx, `SELECT value FROM gauges WHERE name = $1`, name).Scan(&value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || !value.Valid {
			return 0, ErrMetricNotFound
		}
		log.Error("failed to get gauge", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	if !value.Valid {
		return 0, ErrMetricNotFound
	}
	log.Info("got gauge", zap.String("name", name), zap.Float64("value", value.Float64))
	return value.Float64, nil
}

func (ps *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	log := contextkey.LoggerFromContext(ctx)
	var value sql.NullInt64
	err := ps.doWithRetry(ctx, func(ctx context.Context) error {
		return ps.DB.QueryRowContext(ctx, `SELECT value FROM counters WHERE name = $1`, name).Scan(&value)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || !value.Valid {
			return 0, ErrMetricNotFound
		}
		log.Error("failed to get counter", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	if !value.Valid {
		return 0, ErrMetricNotFound
	}
	log.Info("got counter", zap.String("name", name), zap.Int64("value", value.Int64))
	return value.Int64, nil
}

func (ps *PostgresStorage) ListGauges(ctx context.Context) (map[string]float64, error) {
	log := contextkey.LoggerFromContext(ctx)
	rows, err := ps.DB.QueryContext(ctx, "SELECT name, value FROM gauges")
	if err != nil {
		log.Error("failed to query gauges", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value sql.NullFloat64
		if err := rows.Scan(&name, &value); err != nil {
			log.Error("failed to scan gauge row", zap.Error(err))
			continue
		}
		if value.Valid {
			result[name] = value.Float64
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("error iterating gauge rows", zap.Error(err))
	}
	return result, err
}

func (ps *PostgresStorage) ListCounters(ctx context.Context) (map[string]int64, error) {
	log := contextkey.LoggerFromContext(ctx)
	rows, err := ps.DB.QueryContext(ctx, "SELECT name, value FROM counters")
	if err != nil {
		log.Error("failed to query counters", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var name string
		var value sql.NullInt64
		if err := rows.Scan(&name, &value); err != nil {
			log.Error("failed to scan counter row", zap.Error(err))
			continue
		}
		if value.Valid {
			result[name] = value.Int64
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("error iterating counter rows", zap.Error(err))
	}
	return result, err
}

func (ps *PostgresStorage) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	log := contextkey.LoggerFromContext(ctx)
	err := ps.doWithRetry(ctx, func(ctx context.Context) error {
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
	if err != nil {
		log.Error("failed to save all metrics", zap.Error(err))
	}
	return err
}

func (ps *PostgresStorage) CheckStorageAvailability(ctx context.Context) bool {
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := ps.CheckConnection(checkCtx)
	log := contextkey.LoggerFromContext(ctx)
	if err != nil {
		log.Warn("database connection check failed", zap.Error(err))
		return false
	}
	return true
}

func (ps *PostgresStorage) Close() error {
	log := ps.logger
	if ps.DB != nil {
		err := ps.DB.Close()
		if err != nil {
			log.Error("failed to close Postgres DB", zap.Error(err))
			return err
		}
	}
	log.Info("postgres storage closed")
	return nil
}
