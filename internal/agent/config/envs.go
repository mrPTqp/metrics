package config

import (
	"os"
	"strconv"
)

type Envs struct {
	Address        *string
	ReportInterval *int
	PoolInterval   *int
	SecretKey      *string
	RateLimit      *int
}

func ParseEnvs() *Envs {
	var address string
	var reportInterval int
	var poolInterval int
	var secretKey string
	var rateLimit int

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}

	if envReportIntervalStr := os.Getenv("REPORT_INTERVAL"); envReportIntervalStr != "" {
		if envReportInterval, err := strconv.Atoi(envReportIntervalStr); err != nil {
			panic("wrong STORE_INTERVAL type value: " + envReportIntervalStr)
		} else {
			reportInterval = envReportInterval
		}
	}

	if envPoolIntervalStr := os.Getenv("POLL_INTERVAL"); envPoolIntervalStr != "" {
		if poolInterval, err := strconv.Atoi(envPoolIntervalStr); err != nil {
			panic("wrong STORE_INTERVAL type value: " + envPoolIntervalStr)
		} else {
			reportInterval = poolInterval
		}
	}

	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = envSecretKey
	}

	if envRateLimitStr := os.Getenv("RATE_LIMIT"); envRateLimitStr != "" {
		if envRateLimit, err := strconv.Atoi(envRateLimitStr); err != nil {
			panic("wrong STORE_INTERVAL type value: " + envRateLimitStr)
		} else {
			rateLimit = envRateLimit
		}
	}

	return &Envs{
		Address:        &address,
		ReportInterval: &reportInterval,
		PoolInterval:   &poolInterval,
		SecretKey:      &secretKey,
		RateLimit:      &rateLimit,
	}
}
