package main

import (
	"errors"
	"strconv"
	"strings"
)

type Config struct {
	Address        NetAddress
	ReportInterval int
	PoolInterval   int
}

type NetAddress struct {
	Host string
	Port int
}

func LoadConfig() *Config {
	config := Config{}

	flags := parseFlags()
	envs := parseEnvs()

	na := NetAddress{}
	address := "localhost:8080"
	if envs.Address != "" {
		address = envs.Address
	} else if flags.Address != "" {
		address = flags.Address
	}
	if err := na.setAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}	
	config.Address = na

	if envs.ReportInterval != 0 {
		config.ReportInterval = envs.ReportInterval
	} else if flags.ReportInterval != 0 {
		config.ReportInterval = flags.ReportInterval
	} else {
		config.ReportInterval = 10
	}

	if envs.PoolInterval != 0 {
		config.PoolInterval = envs.PoolInterval
	} else if flags.PoolInterval != 0 {
		config.PoolInterval = flags.PoolInterval
	} else {
		config.PoolInterval = 2
	}

	return &config
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
