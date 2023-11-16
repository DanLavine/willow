package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type LimiterConfig struct {
	// logger config
	logLevel *string

	// server config
	LimiterPort *string

	// server certs
	LimiterCA        *string
	LimiterServerKey *string
	LimiterServerCRT *string

	// certificates for locker client
	LockerURL       *string
	LockerClientCA  *string
	LockerClientKey *string
	LockerClientCRT *string

	// use http instead of https
	LimiterInsecureHTTP *bool
}

func Limiter(args []string) (*LimiterConfig, error) {
	limiterFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	limiterFlagSet.Usage = func() {
		fmt.Printf(`Limiter usage:
All flags will use the env vars if they are set instead of command line parameters.

`)
		limiterFlagSet.PrintDefaults()
	}

	limiterConfig := &LimiterConfig{
		logLevel: limiterFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL"),

		LimiterPort: limiterFlagSet.String("limiter-port", "8082", "default port for the limitter server to run on. Can be set by env var LIMITER_PORT"),

		LimiterCA:        limiterFlagSet.String("limiter-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var LIMITER_CA"),
		LimiterServerKey: limiterFlagSet.String("limiter-server-key", "", "Server private key location on disk. Can be set by env var LIMITER_SERVER_KEY"),
		LimiterServerCRT: limiterFlagSet.String("limiter-server-crt", "", "Server ssl certificate location on disk. Can be st by env var LIMITER_SERVER_CRT"),

		LockerURL:       limiterFlagSet.String("limiter-locker-url", "", "CA file used to generate server certs iff one was used. Can be set by env var LIMITER_LOCKER_URL"),
		LockerClientCA:  limiterFlagSet.String("limiter-locker-client-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var LIMITER_LOCKER_CLIENT_CA"),
		LockerClientKey: limiterFlagSet.String("limiter-locker-client-key", "", "Client private key location on disk. Can be set by env var LIMITER_LOCKER_CLIENT_KEY"),
		LockerClientCRT: limiterFlagSet.String("limiter-locker-client-crt", "", "Client ssl certificate location on disk. Can be st by env var LIMITER_LOCKER_CIENT_CRT"),

		LimiterInsecureHTTP: limiterFlagSet.Bool("limiter-insecure-http", false, "Can be used to run the server in an unsecure http mode. Can be set be env var LIMITER_INSECURE_HTTP"),
	}

	if err := limiterFlagSet.Parse(args); err != nil {
		return nil, err
	}

	limiterConfig.parseEnv()

	if err := limiterConfig.validate(); err != nil {
		return nil, err
	}

	return limiterConfig, nil
}

func (lc *LimiterConfig) parseEnv() {
	// logs
	//// logLevel
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		lc.logLevel = &logLevel
	}

	// server
	//// port
	if LimiterPort := os.Getenv("LIMITER_PORT"); LimiterPort != "" {
		lc.LimiterPort = &LimiterPort
	}

	// server keys
	//// ca key
	if limiterCA := os.Getenv("LIMITER_CA"); limiterCA != "" {
		lc.LimiterCA = &limiterCA
	}
	//// tls key
	if limiterServerKey := os.Getenv("LIMITER_SERVER_KEY"); limiterServerKey != "" {
		lc.LimiterServerKey = &limiterServerKey
	}
	//// tls certificate
	if limiterServerCRT := os.Getenv("LIMITER_SERVER_CRT"); limiterServerCRT != "" {
		lc.LimiterServerCRT = &limiterServerCRT
	}

	// locker client
	//// url
	if lockerURL := os.Getenv("LIMITER_LOCKER_URL"); lockerURL != "" {
		lc.LockerURL = &lockerURL
	}

	//// http
	//// ca key
	if lockerCA := os.Getenv("LIMITER_LOCKER_CA"); lockerCA != "" {
		lc.LockerClientCA = &lockerCA
	}
	//// tls key
	if lockerKey := os.Getenv("LIMITER_LOCKER_CLIENT_KEY"); lockerKey != "" {
		lc.LockerClientKey = &lockerKey
	}
	//// tls certificate
	if lockerCRT := os.Getenv("LIMITER_LOCKER_CLIENT_CRT"); lockerCRT != "" {
		lc.LockerClientCRT = &lockerCRT
	}

	// Insecure settings
	//// http
	if limiterInsecureHTTP := os.Getenv("LIMITER_INSECURE_HTTP"); limiterInsecureHTTP != "" {
		if strings.ToLower(limiterInsecureHTTP) == "true" {
			trueValue := true
			lc.LimiterInsecureHTTP = &trueValue
		}
	}
}

func (lc *LimiterConfig) validate() error {
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
	LimiterPort, err := strconv.Atoi(*lc.LimiterPort)
	if err != nil {
		return fmt.Errorf("error parsing 'limiter-port': %w", err)
	} else if LimiterPort > 65535 {
		return fmt.Errorf("param 'limiter-port' is invalid: '%d'. Must be set to a proper port below 65536", LimiterPort)
	}

	if *lc.LimiterInsecureHTTP {
		if *lc.LimiterCA != "" {
			return fmt.Errorf("param 'limiter-ca' can not be set with 'limiter-insecure-http'")
		}

		if *lc.LimiterServerCRT != "" {
			return fmt.Errorf("param 'limiter-server-crt' can not be set with 'limiter-insecure-http'")
		}

		if *lc.LimiterServerKey != "" {
			return fmt.Errorf("param 'limiter-server-key' can not be set with 'limiter-insecure-http'")
		}
	} else {
		// ca key is optional. Could be added on a system level

		// tls key
		if *lc.LimiterServerKey == "" {
			return fmt.Errorf("param 'limiter-server-key' is not set")
		}

		// tls certificate
		if *lc.LimiterServerCRT == "" {
			return fmt.Errorf("param 'limiter-server-crt' is not set")
		}
	}

	// clients
	//// nothing to do here. that will be validated on the client's config

	return nil
}

func (lc *LimiterConfig) LogLevel() string {
	return *lc.logLevel
}
