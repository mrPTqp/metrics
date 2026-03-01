package config

import (
	"encoding/json"
	"os"
)

type JSONConfig struct {
	Address         *string `json:"address"`
	StoreInterval   *int    `json:"store_interval"`
	FileStoragePath *string `json:"file_storage_path,omitempty"`
	Restore         *bool   `json:"restore,omitempty"`
	DatabaseDsn     *string `json:"database_dsn,omitempty"`
	SecretKey       *string `json:"secret_key,omitempty"`
	AuditFilePath   *string `json:"audit_file_path,omitempty"`
	AuditURL        *string `json:"audit_url,omitempty"`
	PrivateKeyPath  *string `json:"private_key_path,omitempty"`
}

func ParseJSONConfig(envPath, flagPath *string) *JSONConfig {
	var path *string
	if envPath != nil && *envPath != "" {
		path = envPath
	}
	if flagPath != nil && *flagPath != "" {
		path = flagPath
	}
	if path == nil {
		return nil
	}

	data, err := os.ReadFile(*path)
	if err != nil {
		panic("error parse config file. Wrong path or file corrupted")
	}

	var cfg JSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic("error parse config file. Unmarshal problem :" + err.Error())
	}

	return &cfg
}
