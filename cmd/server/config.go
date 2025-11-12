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
}

func LoadConfig() *Config {
	flags := parseFlags()
	envs := parseEnvs()

	na := models.NetAddress{}
	address := "localhost:8080"
	if envs.Address != "" {
		address = envs.Address
	} else if flags.Address != "" {
		address = flags.Address
	}
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	storeInterval := 300
	if envs.StoreInterval != 0 {
		storeInterval = envs.StoreInterval
	} else if flags.StoreInterval != 0 {
		storeInterval = flags.StoreInterval
	}

	var syncBackupToFile = false
	if storeInterval == 0 {
		syncBackupToFile = true
	}

	fileStoragePath := os.TempDir() + "/"
	if envs.FileStoragePath != "" {
		fileStoragePath = envs.FileStoragePath
	} else if flags.FileStoragePath != "" {
		fileStoragePath = flags.FileStoragePath
	}

	var restore = false
	if envs.Restore {
		restore = envs.Restore
	} else if flags.Restore {
		restore = flags.Restore
	}

	return &Config{
		Address:       na,
		StoreInterval: storeInterval,
		Restore:       restore,
		File:          filepath.FromSlash(fileStoragePath + "events.log"),
		SyncBackupToFile: syncBackupToFile,
	}
}