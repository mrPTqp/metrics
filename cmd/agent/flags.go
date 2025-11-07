package main

import (
	"flag"
)

type Flags struct {
	Address        string
	ReportInterval int
	PoolInterval   int
}

func parseFlags() *Flags {
	addr := flag.String("a", "", "address and port to run server")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	poolInterval := flag.Int( "p", 2, "pool interval in seconds")

	flag.Parse()

	return &Flags{
		Address: *addr,
		ReportInterval: *reportInterval,
		PoolInterval: *poolInterval,
	}
}
