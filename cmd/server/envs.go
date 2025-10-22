package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

func (na *NetAddress) SetEnvAddress(envAddr string) error {
	parts := strings.Split(envAddr, ":")
	if len(parts) == 2 {
		na.Host = parts[0]
		na.Port, _ = strconv.Atoi(parts[1])
	} else {
		return errors.New("wrong env address format")
	}
	return nil
}

func parseEnvs() {
	var envAddr string
	if envAddr = os.Getenv("ADDRESS"); envAddr != "" {
		address.SetEnvAddress(envAddr)
	}
}
