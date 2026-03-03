package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type rawJSONConfig struct {
	Address         *string `json:"address"`
	StoreInterval   *string `json:"store_interval"`
	FileStoragePath *string `json:"store_file,omitempty"`
	Restore         *bool   `json:"restore,omitempty"`
	DatabaseDsn     *string `json:"database_dsn,omitempty"`
	SecretKey       *string `json:"secret_key,omitempty"`
	AuditFilePath   *string `json:"audit_file_path,omitempty"`
	AuditURL        *string `json:"audit_url,omitempty"`
	PrivateKeyPath  *string `json:"crypto_key,omitempty"`
}

type JSONConfig struct {
	Address         *string
	StoreInterval   *int
	FileStoragePath *string
	Restore         *bool
	DatabaseDsn     *string
	SecretKey       *string
	AuditFilePath   *string
	AuditURL        *string
	PrivateKeyPath  *string
}

func ParseJSONConfig(envPath, flagPath *string) (*JSONConfig, error) {
	var path *string
	if envPath != nil && *envPath != "" {
		path = envPath
	}
	if flagPath != nil && *flagPath != "" {
		path = flagPath
	}
	if path == nil {
		return nil, fmt.Errorf("wrong envPath and flagPath")
	}

	data, err := os.ReadFile(*path)
	if err != nil {
		return nil, err
	}

	var raw rawJSONConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	storeIntervalSec, err := parseDurationPtr(raw.StoreInterval)
	if err != nil {
		return nil, err
	}

	return &JSONConfig{
		Address: raw.Address,
		StoreInterval: storeIntervalSec,
		FileStoragePath: raw.FileStoragePath,
		Restore: raw.Restore,
		DatabaseDsn: raw.DatabaseDsn,
		SecretKey: raw.SecretKey,
		AuditFilePath: raw.AuditFilePath,
		AuditURL: raw.AuditURL,
		PrivateKeyPath: raw.PrivateKeyPath,
	}, nil
}

func parseDurationPtr(s *string) (*int, error) {
	if s == nil || *s == "" {
		return nil, fmt.Errorf("bad duration value")
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return nil, err
	}
	secs := int(d.Seconds())
	return &secs, nil
}