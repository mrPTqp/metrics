package config

import (
	"github.com/mrPTqp/metrics/internal/models"
)

type Config struct {
	Address           models.NetAddress
	GRPCServerAddress models.NetAddress
	ReportInterval    int
	PollInterval      int
	SecretKey         *string
	RateLimit         int
	CertPath          *string
	GRPCEnabled       bool
}

// Загрузка конфигурации
func LoadConfig() *Config {
	flags := ParseFlags()
	envs := ParseEnvs()
	jsonConfig, err := ParseJSONConfig(envs.JSONConfigPath, flags.JSONConfigPath)
	if err != nil {
		panic("error parsing json config: " + err.Error())
	}

	na := models.NetAddress{}
	address := pickValue(envs.Address, flags.Address, jsonConfig.Address, "localhost:8080")
	if err := na.SetAddress(address); err != nil {
		panic("invalid HTTP server address: " + address + " error: " + err.Error())
	}

	grpcNa := models.NetAddress{}
	grpcServerAddress := pickValue(envs.GRPCServerAddress, flags.GRPCServerAddress, jsonConfig.GRPCServerAddress, "localhost:8081")
	if err := grpcNa.SetAddress(grpcServerAddress); err != nil {
		panic("invalid gRPC server address: " + grpcServerAddress + " error: " + err.Error())
	}

	reportInterval := pickValue(envs.ReportInterval, flags.ReportInterval, jsonConfig.ReportInterval, 10)
	pollInterval := pickValue(envs.PollInterval, flags.PollInterval, jsonConfig.PollInterval, 2)
	secretKey := pickValue(envs.SecretKey, flags.SecretKey, jsonConfig.SecretKey, "")
	rateLimit := pickValue(envs.RateLimit, flags.RateLimit, jsonConfig.RateLimit, 20)
	certPath := pickValue(envs.CertPath, flags.CertPath, jsonConfig.CertPath, "")

	grpcEnabled := pickValue(envs.GRPCEnabled, flags.GRPCEnabled, jsonConfig.GRPCEnabled, false)

	return &Config{
		Address:           na,
		GRPCServerAddress: grpcNa,
		ReportInterval:    reportInterval,
		PollInterval:      pollInterval,
		SecretKey:         &secretKey,
		RateLimit:         rateLimit,
		CertPath:          &certPath,
		GRPCEnabled:       grpcEnabled,
	}
}

func pickValue[T comparable](env, flag, json *T, def T) T {
	var zero T
	if env != nil && *env != zero {
		return *env
	}
	if flag != nil && *flag != zero {
		return *flag
	}
	if json != nil && *json != zero {
		return *json
	}
	return def
}
