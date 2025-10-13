package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return fmt.Sprint(na.Host + ":" + strconv.Itoa(na.Port))
}

func (na *NetAddress) Set(flagValue string) error {
	parts := strings.Split(flagValue, ":")
	if len(parts) == 2 {
		na.Host = parts[0]
		na.Port, _ = strconv.Atoi(parts[1])
	} else {
		return errors.New("wrong flag a format")
	}
	return nil
}

var address NetAddress = NetAddress{"localhost", 8080}

func parseFlags() {
	flag.Var(&address, "a", "address and port to run server")

	flag.Parse()
}
