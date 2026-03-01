package config

import (
	"flag"
)

type Flags struct {
	Address        *string
	ReportInterval *int
	PollInterval   *int
	SecretKey      *string
	RateLimit      *int
	CertPath       *string
	JSONConfigPath *string
}

// Парсинг флагов
func ParseFlags() *Flags {
	var addr string
	var reportInterval int
	var pollInterval int
	var secretKey string
	var rateLimit int
	var certPath string
	var jsonConfigPath string

	flag.StringVar(&addr, "a", "", "address and port to send metrics")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&secretKey, "k", "", "sign secret key")
	flag.IntVar(&rateLimit, "l", 20, "agent rate limit")
	flag.StringVar(&certPath, "crypto-key", "", "certificate path")
	flag.StringVar(&jsonConfigPath, "c", "", "config file path")
	flag.StringVar(&jsonConfigPath, "config", "", "config file path")

	flag.Parse()

	return &Flags{
		Address:        &addr,
		ReportInterval: &reportInterval,
		PollInterval:   &pollInterval,
		SecretKey:      &secretKey,
		RateLimit:      &rateLimit,
		CertPath:       &certPath,
		JSONConfigPath: &jsonConfigPath,
	}
}
