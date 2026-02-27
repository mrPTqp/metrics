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
	CertPath       *string
}

// Загрузка конфигурации
func LoadConfig() *Config {
	flags := ParseFlags()
	envs := ParseEnvs()

	na := models.NetAddress{}
	address := pickValue(envs.Address, flags.Address, "localhost:8080")
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	reportInterval := pickValue(envs.ReportInterval, flags.ReportInterval, 10)
	pollInterval := pickValue(envs.PollInterval, flags.PollInterval, 2)
	secretKey := pickValue(envs.SecretKey, flags.SecretKey, "")
	rateLimit := pickValue(envs.RateLimit, flags.RateLimit, 20)
	certPath := pickValue(envs.CertPath, flags.CertPath, "")

	return &Config{
		Address:        na,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		SecretKey:      &secretKey,
		RateLimit:      rateLimit,
		CertPath:       &certPath,
	}
}

func pickValue[T comparable](env, flag *T, def T) T {
	var zero T
	if env != nil && *env != zero {
		return *env
	}
	if flag != nil && *flag != zero {
		return *flag
	}
	return def
}
