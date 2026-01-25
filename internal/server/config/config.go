package config

import (
	"os"
	"path/filepath"

	"github.com/mrPTqp/metrics/internal/models"
)

type Config struct {
	Address          models.NetAddress
	StoreInterval    int
	Restore          bool
	File             string
	SyncBackupToFile bool
	DatabaseDsn      *string
	SecretKey        *string
}

func LoadConfig() *Config {
	envs := ParseEnvs()
	flags := ParseFlags()

	na := models.NetAddress{}
	address := pickValue(envs.Address, flags.Address, "localhost:8080")
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}
	
	storeInterval := pickValue(envs.StoreInterval, flags.StoreInterval, 300)
	var syncBackupToFile = false
	if storeInterval == 0 {
		syncBackupToFile = true
	}

	fileStoragePath := pickValue(envs.FileStoragePath, flags.FileStoragePath, os.TempDir() + "/")
	restore := pickValue(envs.Restore, flags.Restore, false)
	databaseDsn := pickValue(envs.DatabaseDsn, flags.DatabaseDsn, "")
	secretKey := pickValue(envs.SecretKey, flags.SecretKey, "")

	return &Config{
		Address:          na,
		StoreInterval:    storeInterval,
		Restore:          restore,
		File:             filepath.FromSlash(fileStoragePath + "events.log"),
		SyncBackupToFile: syncBackupToFile,
		DatabaseDsn:      &databaseDsn,
		SecretKey:        &secretKey,
	}
}

func pickValue[T comparable](env, flag *T, def T) T {
	var zero T
	if env != nil && *env != zero {
		return *env
	}
	if flag != nil && *flag != zero {
		return *flag
	}
	return def
}
