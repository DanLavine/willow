package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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
	WillowPort      *string
	WillowCA        *string
	WillowServerKey *string
	WillowServerCRT *string

	// metrics server coniguration
	MetricsPort *string

	// global queue configurations
	QueueConfig *QueueConfig

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
	Disk *StorageDisk
}

// Disk Storage Configuration
type StorageDisk struct {
	// root storage directory where all message busses will persist data to
	StorageDir *string
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
		logLevel:        willowFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL"),
		WillowPort:      willowFlagSet.String("willow-port", "8080", "default port for the Willow server to run on. Can be set by env var WILLOW_PORT"),
		MetricsPort:     willowFlagSet.String("metrics-port", "8081", "default port for the Willow server to run on. Can be set by env var WILLOW_PORT"),
		WillowCA:        willowFlagSet.String("willow-ca", "", "CA file used to generate server certs iff one was used. Can be set by env var WILLOW_CA"),
		WillowServerKey: willowFlagSet.String("willow-server-key", "", "Server private key location on disk. Can be set by env var WILLOW_SERVER_KEY"),
		WillowServerCRT: willowFlagSet.String("willow-server-crt", "", "Server ssl certificate location on disk. Can be st by env var WILLOW_SERVER_CRT"),
		QueueConfig: &QueueConfig{
			MaxSize:           willowFlagSet.Uint64("queue-max-size", 4096, "max size of a qeueue for any nuber of items that can be enqueued at once. This includes any items that need to be retried. Can be set by env var QUEUE_MAX_SIZE"),
			DeadLetterMaxSize: willowFlagSet.Uint64("dead-letter-queue-max-size", 100, "max size of the dead letter qeueue for any nuber of items that can be saved. Can be set by env var DEAD_LETTER_QUEUE_MAX_SIZE"),
		},
		StorageConfig: &StorageConfig{
			Type: willowFlagSet.String("storage-type", "memory", "storage type to use for persistence [disk| memory]. Can be set by env var STORAGE_TYPE"),
			Disk: &StorageDisk{
				StorageDir: willowFlagSet.String("storage-dir", "", "root location on disk where to save storage data. Can be set by env var DISK_STORAGE_DIR"),
			},
		},
	}

	// parse coommand line flags
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
		wc.WillowPort = &willowPort
	}
	//// ca key
	if willowCA := os.Getenv("WILLOW_CA"); willowCA != "" {
		wc.WillowCA = &willowCA
	}
	//// tls key
	if willowServerKey := os.Getenv("WILLOW_SERVER_KEY"); willowServerKey != "" {
		wc.WillowServerKey = &willowServerKey
	}
	//// tls certificate
	if willowServerCRT := os.Getenv("WILLOW_SERVER_CRT"); willowServerCRT != "" {
		wc.WillowServerCRT = &willowServerCRT
	}

	//// max queue size
	if queueMaxSize := os.Getenv("QUEUE_MAX_SIZE"); queueMaxSize != "" {
		maxSize, err := strconv.ParseUint(queueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse QueueMaxSize: %w", err)
		}
		wc.QueueConfig.MaxSize = &maxSize
	}

	//// max dead letter queue size
	if deadLetterQueueMaxSize := os.Getenv("DEAD_LETTER_QUEUE_MAX_SIZE"); deadLetterQueueMaxSize != "" {
		maxSize, err := strconv.ParseUint(deadLetterQueueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse DeadLetterQueueMaxSize: %w", err)
		}
		wc.QueueConfig.DeadLetterMaxSize = &maxSize
	}

	// storage config
	//// storage type
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		wc.StorageConfig.Type = &storageType
	}

	// disk storage configuration
	//// disk root dir
	if diskStorageLocation := os.Getenv("DISK_STORAGE_DIR"); diskStorageLocation != "" {
		wc.StorageConfig.Disk.StorageDir = &diskStorageLocation
	}

	// metrics server
	//// port
	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		wc.MetricsPort = &metricsPort
	}

	return nil
}

func (wc *WillowConfig) validate() error {
	// log
	if !(*wc.logLevel == "debug" || *wc.logLevel == "info") {
		return fmt.Errorf("Expected config 'LogLevel' to be [debug | info]. Received: '%s'", *wc.logLevel)
	}

	// tls key
	if *wc.WillowServerKey == "" {
		return fmt.Errorf("parameter 'willow-server-key' is not set")
	}

	// tls certificate
	if *wc.WillowServerCRT == "" {
		return fmt.Errorf("parameter 'willow-server-crt' is not set")
	}

	// storage
	switch *wc.StorageConfig.Type {
	case DiskStorage:
		if *wc.StorageConfig.Disk.StorageDir == "" {
			return fmt.Errorf("'disk-storage-dir' is required when storage type is 'disk'")
		}
	case MemoryStorage:
		// nothing to do here
	default:
		return fmt.Errorf("invalid storage type selected '%s'. Must be one of [memory | disk]", *wc.StorageConfig.Type)
	}

	return nil
}
