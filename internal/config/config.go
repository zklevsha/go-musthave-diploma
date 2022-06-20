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
}

func GetConfig() ServerConfig {
	var config ServerConfig

	var runAddrF, accuralAddrF, dsnF string
	flag.StringVar(&runAddrF, "a", runAddrDef,
		fmt.Sprintf("server socket (default: %s)", runAddrDef))
	flag.StringVar(&accuralAddrF, "p", accurallAddrDef,
		fmt.Sprintf("accural system adddress (default: %s)", accurallAddrDef))
	flag.StringVar(&dsnF, "d", "",
		"database connection string (postgres://username:password@localhost:5432/database_name)")
	flag.Parse()

	runAddrEnv := os.Getenv("RUN_ADDRESS")
	accuralAddrEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	dsnEnv := os.Getenv("DATABASE_URI")

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
		panic("DNS string is not set. Set it via DATABASE_URI enviroment variable of -d flag")
	}

	return config
}
