package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func parseInterval(env string, flag string) (time.Duration, error) {
	if env != "" && flag != "" {
		i, errEnv := time.ParseDuration(env)
		if errEnv == nil {
			return i, nil
		}
		log.Printf("WARN main failed to convert env var to time.Duration: %s.", errEnv.Error())
		i, errFlag := time.ParseDuration(flag)
		if errFlag == nil {
			return i, nil
		}
		return time.Duration(0),
			fmt.Errorf("failed to convert both env var and flag to time.Duration: errEnv=%s, errFlag=%s", errEnv, errFlag)
	}

	if env != "" {
		i, err := time.ParseDuration(env)
		if err != nil {
			return time.Duration(0),
				fmt.Errorf("failed to parse flag %s: %s", flag, err.Error())
		}
		return i, nil
	}

	if flag != "" {
		i, err := time.ParseDuration(flag)
		if err != nil {
			return time.Duration(0),
				fmt.Errorf("failed to convert env var to time.Duration: %s", err.Error())
		}
		return i, nil
	}

	return time.Duration(0), fmt.Errorf("both flag and env are empty")

}

const runAddrDef = "127.0.0.1:8080"
const accrualAddrDef = "127.0.0.1:8081"
const accrualURLDef = "http://127.0.0.1:8081"
const accrualDelayDef = time.Duration(30 * time.Second)

type ServerConfig struct {
	RunAddr      string
	AccrualURL   string
	AccrualDelay time.Duration
	DSN          string
	Key          string
}

type AccrualConfig struct {
	RunAddr string
}

func GetConfig() ServerConfig {
	var config ServerConfig

	var runAddrF, accrualURLF, dsnF, keyF, accuralDelayF string
	flag.StringVar(&runAddrF, "a", runAddrDef, "server socket")
	flag.StringVar(&accrualURLF, "p", accrualURLDef, "accrual system adddress")
	flag.StringVar(&accuralDelayF, "i", accrualDelayDef.String(),
		"how often server will get order`s status/accrual from accrual system")
	flag.StringVar(&dsnF, "d", "",
		"database connection string (postgres://username:password@localhost:5432/database_name)")
	flag.StringVar(&keyF, "k", "",
		"server key (used for salt user`s passwords and jwt auth)")
	flag.Parse()

	runAddrEnv := os.Getenv("RUN_ADDRESS")
	accrualURLEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	accrualDelayEnv := os.Getenv("ACCRUAL_SYSTEM_DELAY")
	dsnEnv := os.Getenv("DATABASE_URI")
	keyEnv := os.Getenv("KEY")

	// Run address
	if runAddrEnv != "" {
		config.RunAddr = runAddrEnv
	} else {
		config.RunAddr = runAddrF
	}

	// accrual URL
	if accrualURLEnv != "" {
		config.AccrualURL = accrualURLEnv
	} else {
		config.AccrualURL = accrualURLF
	}

	// accrual delay
	accuralDelay, err := parseInterval(accrualDelayEnv, accuralDelayF)
	if err != nil {
		log.Printf("WARN can`t parse storeInterval (env:%s, flag: %s): %s. Default value will be used (%s)",
			accrualDelayEnv, accuralDelayF, err.Error(), accrualDelayDef)
		config.AccrualDelay = accrualDelayDef
	} else {
		config.AccrualDelay = accuralDelay
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

func GetAccrualConfig() AccrualConfig {
	var config AccrualConfig

	var runAddrF string
	flag.StringVar(&runAddrF, "a", accrualAddrDef, "server socket")
	flag.Parse()

	config.RunAddr = runAddrF
	return config
}
