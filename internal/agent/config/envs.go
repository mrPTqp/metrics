package config

import (
	"os"
	"strconv"
)

type Envs struct {
	Address           *string
	GRPCServerAddress *string
	ReportInterval    *int
	PollInterval      *int
	SecretKey         *string
	RateLimit         *int
	CertPath          *string
	JSONConfigPath    *string
	GRPCEnabled       *bool
}

// Парсинг переменных окружения
func ParseEnvs() *Envs {
	var address *string
	var grpcServerAddress *string
	var reportInterval *int
	var pollInterval *int
	var secretKey *string
	var rateLimit *int
	var certPath *string
	var jsonConfigPath *string

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = &envAddr
	}

	if envGRPCAddr := os.Getenv("GRPC_SERVER_ADDRESS"); envGRPCAddr != "" {
		grpcServerAddress = &envGRPCAddr
	}

	if envReportIntervalStr := os.Getenv("REPORT_INTERVAL"); envReportIntervalStr != "" {
		if envReportInterval, err := strconv.Atoi(envReportIntervalStr); err != nil {
			panic("wrong REPORT_INTERVAL type value: " + envReportIntervalStr)
		} else {
			reportInterval = &envReportInterval
		}
	}

	if envPollIntervalStr := os.Getenv("POLL_INTERVAL"); envPollIntervalStr != "" {
		if envPollInterval, err := strconv.Atoi(envPollIntervalStr); err != nil {
			panic("wrong POLL_INTERVAL type value: " + envPollIntervalStr)
		} else {
			pollInterval = &envPollInterval
		}
	}

	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = &envSecretKey
	}

	if envRateLimitStr := os.Getenv("RATE_LIMIT"); envRateLimitStr != "" {
		if envRateLimit, err := strconv.Atoi(envRateLimitStr); err != nil {
			panic("wrong RATE_LIMIT type value: " + envRateLimitStr)
		} else {
			rateLimit = &envRateLimit
		}
	}

	if envCertPath := os.Getenv("CRYPTO_KEY"); envCertPath != "" {
		certPath = &envCertPath
	}

	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		jsonConfigPath = &envConfigPath
	}

	var grpcEnabled *bool
	if envGRPCEnabledStr := os.Getenv("GRPC_ENABLED"); envGRPCEnabledStr != "" {
		if envGRPCEnabled, err := strconv.ParseBool(envGRPCEnabledStr); err != nil {
			panic("wrong GRPC_ENABLED type value: " + envGRPCEnabledStr)
		} else {
			grpcEnabled = &envGRPCEnabled
		}
	}

	return &Envs{
		Address:           address,
		GRPCServerAddress: grpcServerAddress,
		ReportInterval:    reportInterval,
		PollInterval:      pollInterval,
		SecretKey:         secretKey,
		RateLimit:         rateLimit,
		CertPath:          certPath,
		JSONConfigPath:    jsonConfigPath,
		GRPCEnabled:       grpcEnabled,
	}
}
