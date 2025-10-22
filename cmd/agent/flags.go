package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

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

func parseFlags() {
	flag.Var(&address, "a", "address and port to send client requests")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&poolInterval, "p", 2, "pool interval in seconds")

	flag.Parse()
}
