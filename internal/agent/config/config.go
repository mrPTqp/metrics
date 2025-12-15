package config

import (
	"github.com/mrPTqp/metrics/internal/models"
)

type Config struct {
	Address        models.NetAddress
	ReportInterval int
	PoolInterval   int
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

	poolInterval := 2
	if envs.PoolInterval != nil && *envs.PoolInterval != 0 {
		poolInterval = *envs.PoolInterval
	} else if flags.ReportInterval != nil {
		poolInterval = *flags.PoolInterval
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
		PoolInterval:   poolInterval,
		SecretKey:      &secretKey,
		RateLimit:      rateLimit,
	}
}
