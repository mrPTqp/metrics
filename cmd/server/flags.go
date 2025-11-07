package main

import (
	"flag"
)

type Flags struct {
	Address string
}

func parseFlags() *Flags {
	addr := flag.String("a", "", "address and port to run server")

	flag.Parse()

	return &Flags{
		Address: *addr,
	}
}
