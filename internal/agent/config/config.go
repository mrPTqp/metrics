package config

import (
	"github.com/mrPTqp/metrics/internal/models"
)

type Config struct {
	Address        models.NetAddress
	ReportInterval int
	PollInterval   int
	SecretKey      *string
	RateLimit      int
}

func LoadConfig() *Config {
	flags := ParseFlags()
	envs := ParseEnvs()

	na := models.NetAddress{}
	address := "localhost:8080"
	if envs.Address != nil && *envs.Address != "" {
		address = *envs.Address
	} else if flags.Address != nil && *flags.Address != "" {
		address = *flags.Address
	}
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	reportInterval := 10
	if envs.ReportInterval != nil && *envs.ReportInterval != 0 {
		reportInterval = *envs.ReportInterval
	} else if flags.ReportInterval != nil {
		reportInterval = *flags.ReportInterval
	}

	pollInterval := 2
	if envs.PollInterval != nil && *envs.PollInterval != 0 {
		pollInterval = *envs.PollInterval
	} else if flags.ReportInterval != nil {
		pollInterval = *flags.PollInterval
	}

	var secretKey string
	if envs.SecretKey != nil && *envs.SecretKey != "" {
		secretKey = *envs.SecretKey
	} else if flags.SecretKey != nil && *flags.SecretKey != "" {
		secretKey = *flags.SecretKey
	}

	rateLimit := 20
	if envs.RateLimit != nil && *envs.RateLimit != 0 {
		rateLimit = *envs.RateLimit
	} else if flags.RateLimit != nil {
		rateLimit = *flags.RateLimit
	}

	return &Config{
		Address:        na,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		SecretKey:      &secretKey,
		RateLimit:      rateLimit,
	}
}
