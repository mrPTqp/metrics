package main

import (
	"os"
	"strconv"
)

type Envs struct {
	Address        string
	ReportInterval int
	PoolInterval   int
}

func parseEnvs() *Envs {
	var address string
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}
	var reportInterval int
	if envReportIntervalStr := os.Getenv("REPORT_INTERVAL"); envReportIntervalStr != "" {
		envReportInterval, err := strconv.Atoi(envReportIntervalStr)
		if err != nil {
			panic("wrong REPORT_INTERVAL type value: " + envReportIntervalStr)
		}
		reportInterval = envReportInterval
	}
	var poolInterval int
	if envPoolIntervalStr := os.Getenv("POLL_INTERVAL"); envPoolIntervalStr != "" {
		envPoolInterval, err := strconv.Atoi(envPoolIntervalStr)
		if err != nil {
			panic("wrong POLL_INTERVAL type value: " + envPoolIntervalStr)
		}
		poolInterval = envPoolInterval
	}

	return &Envs{
		Address:        address,
		ReportInterval: reportInterval,
		PoolInterval:   poolInterval,
	}
}
