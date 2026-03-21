package config

import (
	"flag"
)

type Flags struct {
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

// Парсинг флагов
func ParseFlags() *Flags {
	var addr string
	var grpcServerAddr string
	var reportInterval int
	var pollInterval int
	var secretKey string
	var rateLimit int
	var certPath string
	var jsonConfigPath string

	flag.StringVar(&addr, "a", "", "address and port to send metrics")
	flag.StringVar(&grpcServerAddr, "grpc-a", "", "gRPC server address and port to send metrics")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&secretKey, "k", "", "sign secret key")
	flag.IntVar(&rateLimit, "l", 20, "agent rate limit")
	flag.StringVar(&certPath, "crypto-key", "", "certificate path")
	flag.StringVar(&jsonConfigPath, "c", "", "config file path")
	flag.StringVar(&jsonConfigPath, "config", "", "config file path")

	var grpcEnabled bool
	flag.BoolVar(&grpcEnabled, "grpc-enabled", false, "enable gRPC client")

	flag.Parse()

	return &Flags{
		Address:           &addr,
		GRPCServerAddress: &grpcServerAddr,
		ReportInterval:    &reportInterval,
		PollInterval:      &pollInterval,
		SecretKey:         &secretKey,
		RateLimit:         &rateLimit,
		CertPath:          &certPath,
		JSONConfigPath:    &jsonConfigPath,
		GRPCEnabled:       &grpcEnabled,
	}
}
