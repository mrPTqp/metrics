package config

import (
	"github.com/mrPTqp/metrics/internal/models"
)

type Config struct {
	Address        models.NetAddress
	ReportInterval int
	PoolInterval   int
}

func LoadConfig() *Config {
	config := Config{}

	flags := ParseFlags()
	envs := ParseEnvs()

	na := models.NetAddress{}
	address := "localhost:8080"
	if envs.Address != "" {
		address = envs.Address
	} else if flags.Address != "" {
		address = flags.Address
	}
	if err := na.SetAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}
	config.Address = na

	if envs.ReportInterval != 0 {
		config.ReportInterval = envs.ReportInterval
	} else if flags.ReportInterval != 0 {
		config.ReportInterval = flags.ReportInterval
	} else {
		config.ReportInterval = 10
	}

	if envs.PoolInterval != 0 {
		config.PoolInterval = envs.PoolInterval
	} else if flags.PoolInterval != 0 {
		config.PoolInterval = flags.PoolInterval
	} else {
		config.PoolInterval = 2
	}

	return &config
}
