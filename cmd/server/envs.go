package main

import (
	"os"
	"strconv"
)

type Envs struct {
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func parseEnvs() *Envs {
	var address string
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}
	var storeInterval int
	if envStoreIntervalStr := os.Getenv("STORE_INTERVAL"); envStoreIntervalStr != "" {
		if envStoreInterval, err := strconv.Atoi(envStoreIntervalStr); err != nil {
			panic("wrong STORE_INTERVAL type value: " + envStoreIntervalStr)
		} else {
			storeInterval = envStoreInterval
		}
	}
	var fileStoragePath string
	if envfileStoragePath := os.Getenv("ADDRESS"); envfileStoragePath != "" {
		fileStoragePath = envfileStoragePath
	}
	var restore bool
	if envRestoreStr := os.Getenv("RESTORE"); envRestoreStr != "" {
		if envRestore, err := strconv.ParseBool(envRestoreStr); err != nil {
			panic("wrong RESTORE type value: " + envRestoreStr)
		} else {
			restore = envRestore
		}
	}
	return &Envs{
		Address:         address,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
	}
}
