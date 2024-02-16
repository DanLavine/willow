package config

import (
	"flag"
	"fmt"
	"os"
)

var (
	DiskStorage    = "disk"
	MemoryStorage  = "memory"
	InvalidStorage = "invalid"
)

type WillowConfig struct {
	// Log Level [debug | info]
	logLevel *string

	// willow server configuration
	InsecureHttp *bool
	Port         *string
	ServerCA     *string
	ServerKey    *string
	ServerCRT    *string

	// certificates for limiter client
	LimiterURL       *string
	LimiterClientCA  *string
	LimiterClientKey *string
	LimiterClientCRT *string

	// global storage configurations
	StorageConfig *StorageConfig
}

// Queue configuration
type QueueConfig struct {
	// max number of size any queue can be configured for
	MaxSize *uint64

	// max size of a dead letter queue.
	DeadLetterMaxSize *uint64
}

// Storage configuration
type StorageConfig struct {
	Type *string
}

func Willow(args []string) (*WillowConfig, error) {
	willowFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	willowFlagSet.Usage = func() {
		fmt.Printf(`Willow usage:
All flags will use the env vars if they are set instead of command line parameters.
`)
		willowFlagSet.PrintDefaults()
	}

	willowConfig := &WillowConfig{
		logLevel:         willowFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL"),
		Port:             willowFlagSet.String("port", "8080", "default port for the Willow server to run on. Can be set by env var WILLOW_PORT"),
		InsecureHttp:     willowFlagSet.Bool("insecure-http", false, "Can be used to run the server in an unsecure http mode. Can be set be env var WILLOW_INSECURE_HTTP"),
		ServerCA:         willowFlagSet.String("server-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var WILLOW_CA"),
		ServerKey:        willowFlagSet.String("server-key", "", "Server private key location on disk. Can be set by env var WILLOW_SERVER_KEY"),
		ServerCRT:        willowFlagSet.String("server-crt", "", "Server ssl certificate location on disk. Can be st by env var WILLOW_SERVER_CRT"),
		LimiterURL:       willowFlagSet.String("limiter-url", "", "CA file used to generate server certs iff one was used. Can be set by env var WILLOW_LIMITER_URL"),
		LimiterClientCA:  willowFlagSet.String("limiter-client-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var WILLOW_LIMITER_CLIENT_CA"),
		LimiterClientKey: willowFlagSet.String("limiter-client-key", "", "Client private key location on disk. Can be set by env var WILLOW_LIMITER_CLIENT_KEY"),
		LimiterClientCRT: willowFlagSet.String("limiter-client-crt", "", "Client ssl certificate location on disk. Can be set by env var WILLOW_LIMITER_CLIENT_CRT"),
		StorageConfig: &StorageConfig{
			Type: willowFlagSet.String("storage-type", "memory", "storage type to use for persistence [memory]. Can be set by env var STORAGE_TYPE"),
		},
	}

	// parse command line flags
	if err := willowFlagSet.Parse(args); err != nil {
		return nil, err
	}

	// parse env var flags
	if err := willowConfig.parseEnv(); err != nil {
		return nil, err
	}

	// validate all flags
	if err := willowConfig.validate(); err != nil {
		return nil, err
	}

	return willowConfig, nil
}

func (wc *WillowConfig) LogLevel() string {
	return *wc.logLevel
}

func (wc *WillowConfig) parseEnv() error {
	// logs
	//// log level
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		wc.logLevel = &logLevel
	}

	// willow server
	//// port
	if willowPort := os.Getenv("WILLOW_PORT"); willowPort != "" {
		wc.Port = &willowPort
	}
	//// insecure http
	if willowInsecureHTTP := os.Getenv("WILLOW_INSECURE_HTTP"); willowInsecureHTTP != "" {
		if willowInsecureHTTP == "true" {
			trueValue := true
			wc.InsecureHttp = &trueValue
		}
	}
	//// ca key
	if willowCA := os.Getenv("WILLOW_CA"); willowCA != "" {
		wc.ServerCA = &willowCA
	}
	//// tls key
	if willowServerKey := os.Getenv("WILLOW_SERVER_KEY"); willowServerKey != "" {
		wc.ServerKey = &willowServerKey
	}
	//// tls certificate
	if willowServerCRT := os.Getenv("WILLOW_SERVER_CRT"); willowServerCRT != "" {
		wc.ServerCRT = &willowServerCRT
	}

	// limiter client
	//// url
	if limiterURL := os.Getenv("WILLOW_LIMITER_URL"); limiterURL != "" {
		wc.LimiterURL = &limiterURL
	}
	//// ca key
	if limiterCA := os.Getenv("WILLOW_LIMITER_CA"); limiterCA != "" {
		wc.LimiterClientCA = &limiterCA
	}
	//// tls key
	if limiterKey := os.Getenv("WILLOW_LIMITER_CLIENT_KEY"); limiterKey != "" {
		wc.LimiterClientKey = &limiterKey
	}
	//// tls certificate
	if limiterCRT := os.Getenv("WILLOW_LIMITER_CLIENT_CRT"); limiterCRT != "" {
		wc.LimiterClientCRT = &limiterCRT
	}

	// storage config
	//// storage type
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		wc.StorageConfig.Type = &storageType
	}

	return nil
}

func (wc *WillowConfig) validate() error {
	// log
	if !(*wc.logLevel == "debug" || *wc.logLevel == "info") {
		return fmt.Errorf("expected config 'LogLevel' to be [debug | info]. Received: '%s'", *wc.logLevel)
	}

	if *wc.InsecureHttp {
		if *wc.ServerCA != "" {
			return fmt.Errorf("flag 'server-ca' cannot be set with 'insecure-http'")
		}

		if *wc.ServerCRT != "" {
			return fmt.Errorf("flag 'server-crt' cannot be set with 'insecure-http'")
		}

		if *wc.ServerKey != "" {
			return fmt.Errorf("flag 'server-key' cannot be set with 'insecure-http'")
		}
	} else {
		// tls key
		if *wc.ServerKey == "" {
			return fmt.Errorf("flag 'server-key' is not set")
		}

		// tls certificate
		if *wc.ServerCRT == "" {
			return fmt.Errorf("flag 'server-crt' is not set")
		}
	}

	// storage
	switch *wc.StorageConfig.Type {
	case MemoryStorage:
		// nothing to do here
	default:
		return fmt.Errorf("invalid storage type selected '%s'. Must be one of [memory | disk]", *wc.StorageConfig.Type)
	}

	return nil
}
