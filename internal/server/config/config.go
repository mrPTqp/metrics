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

	fileStoragePath := pickValue(envs.FileStoragePath, flags.FileStoragePath, os.TempDir()+"/")
	fileStoragePath = filepath.FromSlash(fileStoragePath + "backup.log")
	restore := pickValue(envs.Restore, flags.Restore, false)
	databaseDsn := pickValue(envs.DatabaseDsn, flags.DatabaseDsn, "")
	secretKey := pickValue(envs.SecretKey, flags.SecretKey, "")

	var fileAuditEnabled bool
	auditFilePath := pickValue(envs.AuditFilePath, flags.AuditFilePath, "")
	if auditFilePath != "" {
		fileAuditEnabled = true
		auditFilePath = filepath.FromSlash(auditFilePath + "audit.log")
	}

	auditURL := pickValue(envs.AuditURL, flags.AuditURL, "")

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
	}
}

func pickValue[T comparable](env, flag *T, def T) T {
	if env != nil {
		return *env
	}
	if flag != nil {
		return *flag
	}
	return def
}
