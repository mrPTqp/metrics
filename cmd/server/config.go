package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	Address NetAddress
}

type NetAddress struct {
	Host string
	Port int
}

func LoadConfig() *Config {
	flags := parseFlags()
	envs := parseEnvs()

	na := NetAddress{}
	address := "localhost:8080"
	if envs.Address != "" {
		address = envs.Address
		log.Println("Using address from ENV:", address)
	} else if flags.Address != "" {
		address = flags.Address
		log.Println("Using address from FLAG:", address)
	}
	log.Println("DEBUG: Final address before parsing:", address)
	if err := na.setAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	return &Config{
		Address: na,
	}
}

func (na *NetAddress) setAddress(address string) error {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return errors.New("wrong address format, expected host:port")
	}
	na.Host = parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.New("invalid port number: " + parts[1])
	}
	na.Port = port
	return nil
}

func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}
