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

	// port to run the willow tcp server
	WillowPort *string

	// port to run the metrics http server on
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

func Willow() (*WillowConfig, error) {
	willowFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	willowFlagSet.Usage = func() {
		fmt.Printf(`Willow usage:
All flags will use the env vars if they are set instead of command line parameters.

`)
		willowFlagSet.PrintDefaults()
	}

	willowConfig := &WillowConfig{
		logLevel:    willowFlagSet.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL"),
		WillowPort:  willowFlagSet.String("willow-port", "8080", "default port for the Willow server to run on. Can be set by env var WILLOW_PORT"),
		MetricsPort: willowFlagSet.String("metrics-port", "8081", "default port for the Willow server to run on. Can be set by env var WILLOW_PORT"),
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
	if err := willowFlagSet.Parse(os.Args[1:]); err != nil {
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
	// set server defaults
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		wc.logLevel = &logLevel
	}
	if willowPort := os.Getenv("WILLOW_PORT"); willowPort != "" {
		wc.WillowPort = &willowPort
	}
	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		wc.MetricsPort = &metricsPort
	}

	// set queue defaults
	if queueMaxSize := os.Getenv("QUEUE_MAX_SIZE"); queueMaxSize != "" {
		maxSize, err := strconv.ParseUint(queueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse QUEUE_MAX_SIZE: %w", err)
		}
		wc.QueueConfig.MaxSize = &maxSize
	}
	if deadLetterQueueMaxSize := os.Getenv("DEAD_LETTER_QUEUE_MAX_SIZE"); deadLetterQueueMaxSize != "" {
		maxSize, err := strconv.ParseUint(deadLetterQueueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse DEAD_LETTER_QUEUE_MAX_SIZE: %w", err)
		}
		wc.QueueConfig.DeadLetterMaxSize = &maxSize
	}

	// storage config
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		switch storageType {
		case DiskStorage:
			wc.StorageConfig.Type = &DiskStorage

			if diskStorageLocation := os.Getenv("DISK_STORAGE_DIR"); diskStorageLocation != "" {
				wc.StorageConfig.Disk.StorageDir = &diskStorageLocation
			}
		default:
			wc.StorageConfig.Type = &storageType
		}
	}

	return nil
}

func (wc *WillowConfig) validate() error {
	if !(*wc.logLevel == "debug" || *wc.logLevel == "info") {
		return fmt.Errorf("Expected config 'LogLevel' to be [debug | info]. Received: '%s'", *wc.logLevel)
	}

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
