package config

import (
	"encoding/json"
	"os"
)

type JSONConfig struct {
	Address        *string `json:"address"`
	ReportInterval *int    `json:"report_interval"`
	PollInterval   *int    `json:"poll_interval"`
	SecretKey      *string `json:"secret_key,omitempty"`
	RateLimit      *int    `json:"rate_limit,omitempty"`
	CertPath       *string `json:"cert_path,omitempty"`
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
		panic("error parse config file")
	}

	var cfg JSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic("error parse config file")
	}

	return &cfg
}
