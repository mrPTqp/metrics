package config

import (
	"flag"
)

type Flags struct {
	Address         *string
	StoreInterval   *int
	FileStoragePath *string
	Restore         *bool
	DatabaseDsn     *string
	SecretKey       *string
}

func ParseFlags() *Flags {
	var addr string
	var storeInterval int
	var fileStoragePath string
	var restore bool
	var databaseDsn string
	var secretKey string

	flag.StringVar(&addr, "a", "", "address and port to run server")
	flag.IntVar(&storeInterval, "i", 0, "store interval")
	flag.StringVar(&fileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&restore, "r", false, "restore")
	flag.StringVar(&databaseDsn, "d", "", "databse connect string")
	flag.StringVar(&secretKey, "k", "", "sign secret key")

	flag.Parse()

	return &Flags{
		Address:         &addr,
		StoreInterval:   &storeInterval,
		FileStoragePath: &fileStoragePath,
		Restore:         &restore,
		DatabaseDsn:     &databaseDsn,
		SecretKey:       &secretKey,
	}
}
