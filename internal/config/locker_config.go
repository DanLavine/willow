package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type LockerConfig struct {
	// logger config
	logLevel *string

	// server config
	LockerPort      *string
	LockerCA        *string
	LockerServerKey *string
	LockerServerCRT *string
}

func Locker(args []string) (*LockerConfig, error) {
	LockerFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	LockerFlagSet.Usage = func() {
		fmt.Printf(`Locker usage:
All flags will use the env vars if they are set instead of command line parameters.

`)
		LockerFlagSet.PrintDefaults()
	}

	LockerConfig := &LockerConfig{
		logLevel:        LockerFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL"),
		LockerPort:      LockerFlagSet.String("Locker-port", "8082", "default port for the limitter server to run on. Can be set by env var Locker_PORT"),
		LockerCA:        LockerFlagSet.String("Locker-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var Locker_CA"),
		LockerServerKey: LockerFlagSet.String("Locker-server-key", "", "Server private key location on disk. Can be set by env var Locker_SERVER_KEY"),
		LockerServerCRT: LockerFlagSet.String("Locker-server-crt", "", "Server ssl certificate location on disk. Can be st by env var Locker_SERVER_CRT"),
	}

	if err := LockerFlagSet.Parse(args); err != nil {
		return nil, err
	}

	LockerConfig.parseEnv()

	if err := LockerConfig.validate(); err != nil {
		return nil, err
	}

	return LockerConfig, nil
}

func (lc *LockerConfig) parseEnv() {
	// logs
	//// logLevel
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		lc.logLevel = &logLevel
	}

	// server
	//// port
	if LockerPort := os.Getenv("Locker_PORT"); LockerPort != "" {
		lc.LockerPort = &LockerPort
	}

	// ca key
	if LockerCA := os.Getenv("Locker_CA"); LockerCA != "" {
		lc.LockerCA = &LockerCA
	}
	// tls key
	if LockerServerKey := os.Getenv("Locker_SERVER_KEY"); LockerServerKey != "" {
		lc.LockerServerKey = &LockerServerKey
	}
	// tls certificate
	if LockerServerCRT := os.Getenv("Locker_SERVER_CRT"); LockerServerCRT != "" {
		lc.LockerServerCRT = &LockerServerCRT
	}
}

func (lc *LockerConfig) validate() error {
	// log
	//// logLevel
	switch *lc.logLevel {
	case "debug", "info":
		// noting to do here, these are valid
	default:
		return fmt.Errorf("param 'log-level' is invalid: '%s'. Must be set to [debug | info]", *lc.logLevel)
	}

	// server
	//// server port
	LockerPort, err := strconv.Atoi(*lc.LockerPort)
	if err != nil {
		return fmt.Errorf("error parsing 'Locker-port': %w", err)
	} else if LockerPort > 65535 {
		return fmt.Errorf("param 'Locker-port' is invalid: '%d'. Must be set to a proper port below 65536", LockerPort)
	}

	// tls key
	if *lc.LockerServerKey == "" {
		return fmt.Errorf("param 'Locker-server-key' is not set")
	}

	// tls certificate
	if *lc.LockerServerCRT == "" {
		return fmt.Errorf("param 'Locker-server-crt' is not set")
	}

	return nil
}

func (lc *LockerConfig) LogLevel() string {
	return *lc.logLevel
}
