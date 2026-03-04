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
	BackupFilePath   string
	SyncBackupToFile bool
	DatabaseDsn      *string
	SecretKey        *string
	AuditFilePath    string
	FileAuditEnabled bool
	AuditURL         *string
	HTTPAuditEnabled bool
	PrivateKeyPath   *string
}

// Загрузка конфигурации
func LoadConfig() *Config {
	envs := ParseEnvs()
	flags := ParseFlags()
	jsonConfig, err := ParseJSONConfig(envs.JSONConfigPath, flags.JSONConfigPath)
	if err != nil {
		panic("error parsing json config: " + err.Error())
	}

	na := models.NetAddress{}
	address := pickValue(envs.Address, flags.Address, jsonConfig.Address, "localhost:8080")
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	storeInterval := pickValue(envs.StoreInterval, flags.StoreInterval, jsonConfig.StoreInterval, 300)
	var syncBackupToFile = false
	if storeInterval == 0 {
		syncBackupToFile = true
	}

	fileStoragePath := pickValue(envs.FileStoragePath, flags.FileStoragePath, jsonConfig.FileStoragePath, os.TempDir()+"/backup.log")
	restore := pickValue(envs.Restore, flags.Restore, jsonConfig.Restore, false)
	databaseDsn := pickValue(envs.DatabaseDsn, flags.DatabaseDsn, jsonConfig.DatabaseDsn, "")
	secretKey := pickValue(envs.SecretKey, flags.SecretKey, jsonConfig.SecretKey, "")

	var fileAuditEnabled bool
	auditFilePath := pickValue(envs.AuditFilePath, flags.AuditFilePath, jsonConfig.AuditFilePath, "")
	if auditFilePath != "" {
		fileAuditEnabled = true
		auditFilePath = filepath.FromSlash(auditFilePath + "audit.log")
	}

	auditURL := pickValue(envs.AuditURL, flags.AuditURL, jsonConfig.AuditURL, "")
	privateKeyPath := pickValue(envs.PrivateKeyPath, flags.PrivateKeyPath, jsonConfig.PrivateKeyPath, "")

	return &Config{
		Address:          na,
		StoreInterval:    storeInterval,
		Restore:          restore,
		BackupFilePath:   fileStoragePath,
		SyncBackupToFile: syncBackupToFile,
		DatabaseDsn:      &databaseDsn,
		SecretKey:        &secretKey,
		AuditFilePath:    auditFilePath,
		FileAuditEnabled: fileAuditEnabled,
		AuditURL:         &auditURL,
		HTTPAuditEnabled: auditURL != "",
		PrivateKeyPath:   &privateKeyPath,
	}
}

func pickValue[T comparable](env, flag, json *T, def T) T {
	var zero T
	if env != nil && *env != zero {
		return *env
	}
	if flag != nil && *flag != zero {
		return *flag
	}
	if json != nil && *json != zero {
		return *json
	}
	return def
}
