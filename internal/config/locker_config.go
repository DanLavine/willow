package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type LockerConfig struct {
	// logger config
	logLevel *string

	// server config
	LockerPort *string

	// server certs
	LockerCA        *string
	LockerServerKey *string
	LockerServerCRT *string

	// use http instead of https
	LockerInsecureHTTP *bool

	// lock default
	LockDefaultTimeout *time.Duration
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
		logLevel:           LockerFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL"),
		LockerPort:         LockerFlagSet.String("locker-port", "8083", "default port for the limitter server to run on. Can be set by env var LOCKER_PORT"),
		LockerCA:           LockerFlagSet.String("locker-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var LOCKER_CA"),
		LockerServerKey:    LockerFlagSet.String("locker-server-key", "", "Server private key location on disk. Can be set by env var LOCKER_SERVER_KEY"),
		LockerServerCRT:    LockerFlagSet.String("locker-server-crt", "", "Server ssl certificate location on disk. Can be st by env var LOCKER_SERVER_CRT"),
		LockerInsecureHTTP: LockerFlagSet.Bool("locker-insecure-http", false, "Can be used to run the server in an unsecure http mode. Can be set be env var LOCKER_INSECURE_HTTP"),
		LockDefaultTimeout: LockerFlagSet.Duration("locker-default-timeout", 5*time.Second, "default timeout for locks if it is not provided by the api. Default is 5 seconds. Can be set by env var LOCKER_DEFAULT_TIMEOUT"),
	}

	if err := LockerFlagSet.Parse(args); err != nil {
		return nil, err
	}

	if err := LockerConfig.parseEnv(); err != nil {
		return nil, err
	}

	if err := LockerConfig.validate(); err != nil {
		return nil, err
	}

	return LockerConfig, nil
}

func (lc *LockerConfig) parseEnv() error {
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

	// insecure http
	if LockerInsecureHTTP := os.Getenv("LLOCKER_INSECURE_HTTP"); LockerInsecureHTTP != "" {
		if LockerInsecureHTTP == "true" {
			trueValue := true
			lc.LockerInsecureHTTP = &trueValue
		}
	}

	// defaults
	// lock timeout
	if LockerDefaultTimeout := os.Getenv("LOCKER_DEFAULT_TIMEOUT"); LockerDefaultTimeout != "" {
		lockDefaultTimeout, err := time.ParseDuration(LockerDefaultTimeout)
		if err != nil {
			return fmt.Errorf("error parsing 'LOCKER_DEFAULT_TIMEOUT': %v", err)
		}

		lc.LockDefaultTimeout = &lockDefaultTimeout
	}

	return nil
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
		return fmt.Errorf("error parsing 'locker-port': %w", err)
	} else if LockerPort > 65535 {
		return fmt.Errorf("param 'locker-port' is invalid: '%d'. Must be set to a proper port below 65536", LockerPort)
	}

	if *lc.LockerInsecureHTTP {
		if *lc.LockerServerCRT != "" {
			return fmt.Errorf("pram 'locker-server-crt' needds to be nil if using the 'locker-insecure-http'")
		}

		if *lc.LockerServerKey != "" {
			return fmt.Errorf("pram 'locker-server-key' needds to be nil if using the 'locker-insecure-http'")
		}

		if *lc.LockerServerCRT != "" {
			return fmt.Errorf("pram 'locker-server-crt' needds to be nil if using the 'locker-insecure-http'")
		}
	} else {
		// tls key
		if *lc.LockerServerKey == "" {
			return fmt.Errorf("param 'locker-server-key' is not set")
		}

		// tls certificate
		if *lc.LockerServerCRT == "" {
			return fmt.Errorf("param 'locker-server-crt' is not set")
		}

		// default timeout
		if *lc.LockDefaultTimeout <= 0 {
			return fmt.Errorf("error parsing 'LOCKER_DEFAULT_TIMEOUT', time must be greater than 0")
		}
	}

	return nil
}

func (lc *LockerConfig) LogLevel() string {
	return *lc.logLevel
}
