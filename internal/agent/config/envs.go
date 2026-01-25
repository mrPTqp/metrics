package config

import (
	"os"
	"strconv"
)

type Envs struct {
	Address        *string
	ReportInterval *int
	PollInterval   *int
	SecretKey      *string
	RateLimit      *int
}

func ParseEnvs() *Envs {
	var address string
	var reportInterval int
	var pollInterval int
	var secretKey string
	var rateLimit int

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}

	if envReportIntervalStr := os.Getenv("REPORT_INTERVAL"); envReportIntervalStr != "" {
		if envReportInterval, err := strconv.Atoi(envReportIntervalStr); err != nil {
			panic("wrong REPORT_INTERVAL type value: " + envReportIntervalStr)
		} else {
			reportInterval = envReportInterval
		}
	}

	if envPollIntervalStr := os.Getenv("POLL_INTERVAL"); envPollIntervalStr != "" {
		if envPollInterval, err := strconv.Atoi(envPollIntervalStr); err != nil {
			panic("wrong POLL_INTERVAL type value: " + envPollIntervalStr)
		} else {
			pollInterval = envPollInterval
		}
	}

	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = envSecretKey
	}

	if envRateLimitStr := os.Getenv("RATE_LIMIT"); envRateLimitStr != "" {
		if envRateLimit, err := strconv.Atoi(envRateLimitStr); err != nil {
			panic("wrong RATE_LIMIT type value: " + envRateLimitStr)
		} else {
			rateLimit = envRateLimit
		}
	}

	return &Envs{
		Address:        &address,
		ReportInterval: &reportInterval,
		PollInterval:   &pollInterval,
		SecretKey:      &secretKey,
		RateLimit:      &rateLimit,
	}
}
