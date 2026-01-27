package migrations

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func RunMigrations(dsn string, logger *zap.Logger) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	current, _, _ := m.Version()
	logger.Info("Current migration version", zap.Uint("version", current))

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("No migrations to apply")
			return nil
		}

		logger.Error("Migration error, attempting rollback...", zap.Error(err))
		if rollbackErr := m.Down(); rollbackErr != nil {
			logger.Error("Rollback after migration failure failed", zap.Error(rollbackErr))
		} else {
			logger.Info("Rollback after migration failure succeeded")
		}

		return err
	}

	newVersion, _, _ := m.Version()
	if newVersion > current {
		logger.Info("Successfully migrated", zap.Uint("from", current), zap.Uint("to", newVersion))
	} else {
		logger.Info("No new migrations found")
	}

	return nil
}
