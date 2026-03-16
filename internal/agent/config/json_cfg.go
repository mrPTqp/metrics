package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type rawJSONConfig struct {
	Address           *string `json:"address"`
	GRPCServerAddress *string `json:"grpc_server_address,omitempty"`
	ReportInterval    *string `json:"report_interval"`
	PollInterval      *string `json:"poll_interval"`
	SecretKey         *string `json:"secret_key,omitempty"`
	RateLimit         *int    `json:"rate_limit,omitempty"`
	CertPath          *string `json:"crypto_key,omitempty"`
	GRPCEnabled       *bool   `json:"grpc_enabled,omitempty"`
}

type JSONConfig struct {
	Address           *string
	GRPCServerAddress *string
	ReportInterval    *int
	PollInterval      *int
	SecretKey         *string
	RateLimit         *int
	CertPath          *string
	GRPCEnabled       *bool
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

	reportSec, err := parseDurationPtr(raw.ReportInterval)
	if err != nil {
		return nil, err
	}
	
	pollSec, err := parseDurationPtr(raw.PollInterval)
	if err != nil {
		return nil, err
	}

	return &JSONConfig{
		Address:           raw.Address,
		GRPCServerAddress: raw.GRPCServerAddress,
		ReportInterval:    reportSec,
		PollInterval:      pollSec,
		SecretKey:         raw.SecretKey,
		RateLimit:         raw.RateLimit,
		CertPath:          raw.CertPath,
		GRPCEnabled:       raw.GRPCEnabled,
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
