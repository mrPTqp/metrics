package main

import (
	"errors"
	"strconv"
	"strings"
)

type Config struct {
	Address          NetAddress
	StoreInterval    int
	Restore          bool
	File             string
	SyncBackupToFile bool
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
	} else if flags.Address != "" {
		address = flags.Address
	}
	if err := na.setAddress(address); err != nil {
		panic("invalid address: " + address + " error: " + err.Error())
	}

	storeInterval := 300
	if envs.StoreInterval != 0 {
		storeInterval = envs.StoreInterval
	} else if flags.StoreInterval != 0 {
		storeInterval = flags.StoreInterval
	}

	var syncBackupToFile = false
	if storeInterval == 0 {
		syncBackupToFile = true
	}

	fileStoragePath := "/tmp/"
	if envs.FileStoragePath != "" {
		fileStoragePath = envs.FileStoragePath
	} else if flags.FileStoragePath != "" {
		fileStoragePath = flags.FileStoragePath
	}

	var restore = false
	if envs.Restore {
		restore = envs.Restore
	} else if flags.Restore {
		restore = flags.Restore
	}

	return &Config{
		Address:       na,
		StoreInterval: storeInterval,
		Restore:       restore,
		File:          fileStoragePath + "events.log",
		SyncBackupToFile: syncBackupToFile,
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
