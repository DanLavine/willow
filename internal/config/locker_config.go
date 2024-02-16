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
	Port *string

	// server certs
	ServerCA  *string
	ServerKey *string
	ServerCRT *string

	// use http instead of https
	InsecureHttp *bool

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
		logLevel:           LockerFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOCKER_LOG_LEVEL"),
		Port:               LockerFlagSet.String("port", "8083", "default port for the limitter server to run on. Can be set by env var LOCKER_PORT"),
		ServerCA:           LockerFlagSet.String("server-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var LOCKER_SERVER_CA"),
		ServerKey:          LockerFlagSet.String("server-key", "", "Server private key location on disk. Can be set by env var LOCKER_SERVER_KEY"),
		ServerCRT:          LockerFlagSet.String("server-crt", "", "Server ssl certificate location on disk. Can be st by env var LOCKER_SERVER_CRT"),
		InsecureHttp:       LockerFlagSet.Bool("insecure-http", false, "Can be used to run the server in an unsecure http mode. Can be set be env var LOCKER_INSECURE_HTTP"),
		LockDefaultTimeout: LockerFlagSet.Duration("lock-default-timeout", 5*time.Second, "default timeout for locks if it is not provided by the api. Default is 5 seconds. Can be set by env var LOCKER_LOCK_DEFAULT_TIMEOUT"),
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
	if logLevel := os.Getenv("LOCKER_LOG_LEVEL"); logLevel != "" {
		lc.logLevel = &logLevel
	}

	// server
	//// port
	if lockerPort := os.Getenv("LOCKER_PORT"); lockerPort != "" {
		lc.Port = &lockerPort
	}

	// ca key
	if serverCA := os.Getenv("LOCKER_SERVER_CA"); serverCA != "" {
		lc.ServerCA = &serverCA
	}
	// tls key
	if lockerServerKey := os.Getenv("LOCKER_SERVER_KEY"); lockerServerKey != "" {
		lc.ServerKey = &lockerServerKey
	}
	// tls certificate
	if lockerServerCRT := os.Getenv("LOCKER_SERVER_CRT"); lockerServerCRT != "" {
		lc.ServerCRT = &lockerServerCRT
	}

	// insecure http
	if insecureHTTP := os.Getenv("LOCKER_INSECURE_HTTP"); insecureHTTP != "" {
		if insecureHTTP == "true" {
			trueValue := true
			lc.InsecureHttp = &trueValue
		}
	}

	// defaults
	// lock timeout
	if lockerLockDefaultTimeout := os.Getenv("LOCKER_LOCK_DEFAULT_TIMEOUT"); lockerLockDefaultTimeout != "" {
		defaultTimeout, err := time.ParseDuration(lockerLockDefaultTimeout)
		if err != nil {
			return fmt.Errorf("error parsing 'LOCKER_LOCK_DEFAULT_TIMEOUT': %v", err)
		}

		lc.LockDefaultTimeout = &defaultTimeout
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
		return fmt.Errorf("flag 'log-level' is invalid: '%s'. Must be set to [debug | info]", *lc.logLevel)
	}

	// server
	//// server port
	LockerPort, err := strconv.Atoi(*lc.Port)
	if err != nil {
		return fmt.Errorf("error parsing 'port': %w", err)
	} else if LockerPort > 65535 {
		return fmt.Errorf("flag 'port' is invalid: '%d'. Must be set to a proper port below 65536", LockerPort)
	}

	if *lc.InsecureHttp {
		if *lc.ServerCA != "" {
			return fmt.Errorf("flag 'server-ca' needs to be unset if using 'insecure-http'")
		}

		if *lc.ServerKey != "" {
			return fmt.Errorf("flag 'server-key' needs to be unset if using 'insecure-http'")
		}

		if *lc.ServerCRT != "" {
			return fmt.Errorf("flag 'server-crt' needs to be unset if using 'insecure-http'")
		}
	} else {
		// ca key is optional. Could be added on a system level

		// tls key
		if *lc.ServerKey == "" {
			return fmt.Errorf("flag 'server-key' is not set")
		}

		// tls certificate
		if *lc.ServerCRT == "" {
			return fmt.Errorf("flag 'server-crt' is not set")
		}
	}

	// default timeout
	if *lc.LockDefaultTimeout <= 0 {
		return fmt.Errorf("error parsing 'LOCKER_LOCK_DEFAULT_TIMEOUT', time must be greater than 0")
	}

	return nil
}

func (lc *LockerConfig) LogLevel() string {
	return *lc.logLevel
}
