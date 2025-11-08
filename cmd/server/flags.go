package main

import (
	"flag"
)

type Flags struct {
	Address string
	StoreInterval int
	FileStoragePath string
	Restore bool
}

func parseFlags() *Flags {
	addr := flag.String("a", "", "address and port to run server")
	storeInterval := flag.Int("i", 0, "store interval")
	fileStoragePath := flag.String("f", "", "file storage path")
	restore := flag.Bool("r", false, "restore")

	flag.Parse()

	return &Flags{
		Address: *addr,
		StoreInterval: *storeInterval,
		FileStoragePath: *fileStoragePath,
		Restore: *restore,		
	}
}
