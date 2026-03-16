package config

import (
	"os"
	"strconv"
)

type Envs struct {
	Address         *string
	GRPCAddress     *string
	StoreInterval   *int
	FileStoragePath *string
	Restore         *bool
	DatabaseDsn     *string
	SecretKey       *string
	AuditFilePath   *string
	AuditURL        *string
	PrivateKeyPath  *string
	JSONConfigPath  *string
	TrustedSubnet   *string
	GRPCEnabled     *bool
}

// Парсинг переменных окружения
func ParseEnvs() *Envs {
	var address *string
	var grpcAddress *string
	var storeInterval *int
	var fileStoragePath *string
	var restore *bool
	var databaseDsn *string
	var secretKey *string
	var auditFilePath *string
	var auditURL *string
	var privateKeyPath *string
	var jsonConfigPath *string
	var trustedSubnet *string

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = &envAddr
	}

	if envGRPCAddr := os.Getenv("GRPC_ADDRESS"); envGRPCAddr != "" {
		grpcAddress = &envGRPCAddr
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

	if envPrivateKeyPath := os.Getenv("CRYPTO_KEY"); envPrivateKeyPath != "" {
		privateKeyPath = &envPrivateKeyPath
	}

	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		jsonConfigPath = &envConfigPath
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		trustedSubnet = &envTrustedSubnet
	}

	var grpcEnabled bool
	if envGRPCEnabledStr := os.Getenv("GRPC_ENABLED"); envGRPCEnabledStr != "" {
		if envGRPCEnabled, err := strconv.ParseBool(envGRPCEnabledStr); err != nil {
			panic("wrong GRPC_ENABLED type value: " + envGRPCEnabledStr)
		} else {
			grpcEnabled = envGRPCEnabled
		}
	}

	return &Envs{
		Address:         address,
		GRPCAddress:     grpcAddress,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
		DatabaseDsn:     databaseDsn,
		SecretKey:       secretKey,
		AuditFilePath:   auditFilePath,
		AuditURL:        auditURL,
		PrivateKeyPath:  privateKeyPath,
		JSONConfigPath:  jsonConfigPath,
		TrustedSubnet:   trustedSubnet,
		GRPCEnabled:     &grpcEnabled,
	}
}
