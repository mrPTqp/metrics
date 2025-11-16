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
