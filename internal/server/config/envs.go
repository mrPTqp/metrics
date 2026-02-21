package config

import (
	"os"
	"strconv"
)

type Envs struct {
	Address         *string
	StoreInterval   *int
	FileStoragePath *string
	Restore         *bool
	DatabaseDsn     *string
	SecretKey       *string
	AuditFilePath   *string
	AuditURL        *string
}

// Парсинг переменных окружения
func ParseEnvs() *Envs {
	var address *string
	var storeInterval *int
	var fileStoragePath *string
	var restore *bool
	var databaseDsn *string
	var secretKey *string
	var auditFilePath *string
	var auditURL *string

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = &envAddr
	}

	if envStoreIntervalStr := os.Getenv("STORE_INTERVAL"); envStoreIntervalStr != "" {
		if envStoreInterval, err := strconv.Atoi(envStoreIntervalStr); err != nil {
			panic("wrong STORE_INTERVAL type value: " + envStoreIntervalStr)
		} else {
			storeInterval = &envStoreInterval
		}
	}

	if envfileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envfileStoragePath != "" {
		fileStoragePath = &envfileStoragePath
	}

	if envRestoreStr := os.Getenv("RESTORE"); envRestoreStr != "" {
		if envRestore, err := strconv.ParseBool(envRestoreStr); err != nil {
			panic("wrong RESTORE type value: " + envRestoreStr)
		} else {
			restore = &envRestore
		}
	}

	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		databaseDsn = &envDatabaseDsn
	}

	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = &envSecretKey
	}

	if envAuditFilePath := os.Getenv("AUDIT_FILE"); envAuditFilePath != "" {
		auditFilePath = &envAuditFilePath
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		auditURL = &envAuditURL
	}

	return &Envs{
		Address:         address,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
		DatabaseDsn:     databaseDsn,
		SecretKey:       secretKey,
		AuditFilePath:   auditFilePath,
		AuditURL:        auditURL,
	}
}
