package main

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
	DatabaseDsn      string
}

func LoadConfig() *Config {
	flags := parseFlags()
	envs := parseEnvs()

	na := models.NetAddress{}
	address := "localhost:8080"
	if envs.Address != nil && *envs.Address != "" {
		address = *envs.Address
	} else if flags.Address != nil && *flags.Address != "" {
		address = *flags.Address
	}
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	storeInterval := 300
	if envs.StoreInterval != nil && *envs.StoreInterval != 0 {
		storeInterval = *envs.StoreInterval
	} else if flags.StoreInterval != nil {
		storeInterval = *flags.StoreInterval
	}

	var syncBackupToFile = false
	if storeInterval == 0 {
		syncBackupToFile = true
	}

	fileStoragePath := os.TempDir() + "/"
	if envs.FileStoragePath != nil && *envs.FileStoragePath != "" {
		fileStoragePath = *envs.FileStoragePath
	} else if flags.FileStoragePath != nil && *flags.FileStoragePath != "" {
		fileStoragePath = *flags.FileStoragePath
	}

	var restore = false
	if envs.Restore != nil && *envs.Restore {
		restore = *envs.Restore
	} else if flags.Restore != nil && *flags.Restore {
		restore = *flags.Restore
	}

	var databaseDsn string
	if envs.DatabaseDsn != nil && *envs.DatabaseDsn != "" {
		databaseDsn = *envs.DatabaseDsn
	} else if flags.DatabaseDsn != nil && *flags.DatabaseDsn != "" {
		databaseDsn = *flags.DatabaseDsn
	}

	return &Config{
		Address:          na,
		StoreInterval:    storeInterval,
		Restore:          restore,
		File:             filepath.FromSlash(fileStoragePath + "events.log"),
		SyncBackupToFile: syncBackupToFile,
		DatabaseDsn:      databaseDsn,
	}
}
