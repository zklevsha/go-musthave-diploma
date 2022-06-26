package config

import (
	"flag"
	"fmt"
	"os"
)

const runAddrDef = ":8081"
const accuralAddrDef = "127.0.0.1:8080"

type ServerConfig struct {
	RunAddr     string
	AccuralAddr string
	DSN         string
	Key         string
}

type AccuralConfig struct {
	RunAddr string
}

func GetConfig() ServerConfig {
	var config ServerConfig

	var runAddrF, accuralAddrF, dsnF, keyF string
	flag.StringVar(&runAddrF, "a", runAddrDef,
		fmt.Sprintf("server socket (default: %s)", runAddrDef))
	flag.StringVar(&accuralAddrF, "p", accuralAddrDef,
		fmt.Sprintf("accural system adddress (default: %s)", accuralAddrDef))
	flag.StringVar(&dsnF, "d", "",
		"database connection string (postgres://username:password@localhost:5432/database_name)")
	flag.StringVar(&keyF, "k", "",
		"server key (used for salt user`s passwords and jwt auth)")
	flag.Parse()

	runAddrEnv := os.Getenv("RUN_ADDRESS")
	accuralAddrEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	dsnEnv := os.Getenv("DATABASE_URI")
	keyEnv := os.Getenv("KEY")

	// Run address
	if runAddrEnv != "" {
		config.RunAddr = runAddrEnv
	} else {
		config.RunAddr = runAddrF
	}

	// Accural address
	if accuralAddrEnv != "" {
		config.AccuralAddr = accuralAddrEnv
	} else {
		config.AccuralAddr = accuralAddrF
	}

	// DSN
	if dsnEnv != "" {
		config.DSN = dsnEnv
	} else if dsnF != "" {
		config.DSN = dsnF
	} else {
		panic("DNS string is not set. " +
			"Set it via 'DATABASE_URI' enviroment variable or '-d' flag")
	}

	// Key
	if keyEnv != "" {
		config.Key = keyEnv
	} else if keyF != "" {
		config.Key = keyF
	} else {
		panic("server key is not set. " +
			"Set it via 'KEY' enviroment variable or '-k' flag")
	}

	return config
}

func GetAccuralConfig() AccuralConfig {
	var config AccuralConfig

	var runAddrF string
	flag.StringVar(&runAddrF, "a", accuralAddrDef, "server socket")
	flag.Parse()

	config.RunAddr = runAddrF
	return config
}
