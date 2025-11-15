package main

import (
	"flag"
)

type Flags struct {
	Address *string
	StoreInterval *int
	FileStoragePath *string
	Restore *bool
}

func parseFlags() *Flags {
	var addr string
	var storeInterval int
	var fileStoragePath string
	var restore bool

	flag.StringVar(&addr, "a", "", "address and port to run server")
	flag.IntVar(&storeInterval, "i", 0, "store interval")
	flag.StringVar(&fileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&restore, "r", false, "restore")

	flag.Parse()

	return &Flags{
		Address: &addr,
		StoreInterval: &storeInterval,
		FileStoragePath: &fileStoragePath,
		Restore: &restore,		
	}
}