package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PostgresStorage struct {
	DB     *sql.DB
	logger *zap.SugaredLogger
}

func NewPostgresStorage(databaseDsn string, logger *zap.SugaredLogger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		DB:     db,
		logger: logger,
	}, nil
}

func (ps *PostgresStorage) SaveGauge(name string, value *float64) error {
	_, err := ps.DB.Exec(`
		INSERT INTO gauges (name, value) 
		VALUES ($1, $2) 
		ON CONFLICT (name) 
		DO UPDATE SET value = EXCLUDED.value`,
		name, value)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostgresStorage) SaveCounter(name string, value *int64) error {
	_, err := ps.DB.Exec(`
		INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = counters.value + EXCLUDED.value`,
		name, value)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostgresStorage) GetGauge(name string) (float64, error) {
	row := ps.DB.QueryRowContext(context.Background(),
		"SELECT value FROM gauges WHERE name = $1", name)

	var value float64
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (ps *PostgresStorage) GetCounter(name string) (int64, error) {
	row := ps.DB.QueryRowContext(context.Background(),
		"SELECT value FROM counters WHERE name = $1", name)

	var value int64
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (ps *PostgresStorage) ListGauges() (map[string]float64, error) {
	rows, err := ps.DB.QueryContext(context.Background(),
		"SELECT name, value FROM gauges")
	if err != nil {
		return nil, err
	}

	gauges := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64

		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		gauges[name] = value
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return gauges, nil
}
func (ps *PostgresStorage) ListCounters() (map[string]int64, error) {
	rows, err := ps.DB.QueryContext(context.Background(),
		"SELECT name, value FROM counters")
	if err != nil {
		return nil, err
	}

	counters := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64

		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		counters[name] = value
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return counters, nil
}

func (ps *PostgresStorage) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	for name, value := range gauges {
		if err := ps.SaveGauge(name, &value); err != nil {
			return err
		}
	}

	for name, value := range counters {
		if err := ps.SaveCounter(name, &value); err != nil {
			return err
		}
	}

	return nil
}

func (ps *PostgresStorage) CheckStorageAvailability() bool {
	err := ps.CheckConnection()
	return err == nil
}

func (ps *PostgresStorage) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := ps.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
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
