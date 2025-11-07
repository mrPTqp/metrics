package main

import (
	"os"
)

type Envs struct {
	Address string
}

func parseEnvs() *Envs {
	var address string
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}

	return &Envs{
		Address: address,
	}
}
