package config

import (
	"flag"
)

type Flags struct {
	Address        *string
	ReportInterval *int
	PoolInterval   *int
	SecretKey      *string
	RateLimit      *int
}

func ParseFlags() *Flags {
	var addr string
	var reportInterval int
	var poolInterval int
	var secretKey string
	var rateLimit int

	flag.StringVar(&addr, "a", "", "address and port to send metrics")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&poolInterval, "p", 2, "pool interval in seconds")
	flag.StringVar(&secretKey, "k", "", "sign secret key")
	flag.IntVar(&rateLimit, "l", 20, "agent rate limit")

	flag.Parse()

	return &Flags{
		Address:        &addr,
		ReportInterval: &reportInterval,
		PoolInterval:   &poolInterval,
		SecretKey:      &secretKey,
		RateLimit:      &rateLimit,
	}
}
