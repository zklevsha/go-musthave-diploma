package config

import (
	"flag"
	"fmt"
	"os"
)

const runAddrDef = ":8081"
const accurallAddrDef = "127.0.0.1:8080"

type ServerConfig struct {
	RunAddr      string
	AccurallAddr string
	DSN          string
	Key          string
}

func GetConfig() ServerConfig {
	var config ServerConfig

	var runAddrF, accuralAddrF, dsnF, keyF string
	flag.StringVar(&runAddrF, "a", runAddrDef,
		fmt.Sprintf("server socket (default: %s)", runAddrDef))
	flag.StringVar(&accuralAddrF, "p", accurallAddrDef,
		fmt.Sprintf("accural system adddress (default: %s)", accurallAddrDef))
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
		config.AccurallAddr = accuralAddrEnv
	} else {
		config.AccurallAddr = accuralAddrF
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
