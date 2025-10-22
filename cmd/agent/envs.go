package main

import (
	"errors"
	"log"
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
	if envReportIntervalStr := os.Getenv("REPORT_INTERVAL"); envReportIntervalStr != "" {
		envReportInterval, err := strconv.Atoi(envReportIntervalStr)
		if err != nil {
			log.Printf("[ERROR] wrong REPORT_INTERVAL value: %s", envReportIntervalStr)
		}
		reportInterval = envReportInterval
	}
	if envPoolIntervalStr := os.Getenv("POLL_INTERVAL"); envPoolIntervalStr != "" {
		envPoolInterval, err := strconv.Atoi(envPoolIntervalStr)
		if err != nil {
			log.Printf("[ERROR] wrong POLL_INTERVAL value: %s", envPoolIntervalStr)
		}
		poolInterval = envPoolInterval
	}
}
