package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	logLevel    = flag.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL")
	willowPort  = flag.String("willow-port", "8080", "willow server port. Can be set by env var WILLOW_PORT")
	metricsPort = flag.String("metrics-port", "8081", "willow server metrics port. can be set by env var METRICS_PORT")

	// general queue configurations
	queueMaxSize     = flag.Uint64("queue-max-size", 4096, "max size of a qeueue for any nuber of items that can be enqueued at once. This includes any items that need to be retried. Can be set by env var QUEUE_MAX_SIZE")
	queueDefaultSize = flag.Uint64("queue-defaut-size", 2048, "default size of a qeueue for any nuber of items that can be enqueued at once. This includes any items that need to be retried. Can be set by env var QUEUE_DEFAULT_SIZE")

	// there is no default here. If these are not configured, then they are not provided
	deadLetterQueueMaxSize = flag.Uint64("dead-letter-queue-max-size", 100, "max size of the dead letter qeueue for any nuber of items that can be saved. Can be set by env var DEAD_LETTER_QUEUE_MAX_SIZE")

	// storage configurations
	storageType = flag.String("storage-type", "disk", "storage type to use for persistence [disk| memory]. Can be set by env var STORAGE_TYPE")
	// disck storage configurations
	diskStorageDir = flag.String("disk-storage-dir", "", "root location on disk where to save storage data. Can be set by env var DISK_STORAGE_DIR")
)

type StorageType string

const (
	DiskStorage    StorageType = "disk"
	MemoryStorage  StorageType = "memory"
	InvalidStorage StorageType = "invalid"
)

type Config struct {
	// Log Level [debug | info]
	LogLevel string

	// port to run the willow tcp server
	WillowPort string

	// port to run the metrics http server on
	MetricsPort string

	// global queue configurations
	QueueConfig *QueueConfig

	// global storage configurations
	StorageConfig *StorageConfig
}

// Queue configuration
type QueueConfig struct {
	// max number of size any queue can be configured for
	MaxSize uint64
	// Default size for any queue to be configured if no size was provided
	DefaultSize uint64

	// max size of a dead letter queue.
	DeadLetterMaxSize uint64
}

// Storage configuration
type StorageConfig struct {
	Type StorageType
	Disk *StorageDisk
}

// Disk Storage Configuration
type StorageDisk struct {
	// root storage directory where all message busses will persist data to
	StorageDir string
}

func Default() *Config {
	return &Config{
		QueueConfig: &QueueConfig{},
		StorageConfig: &StorageConfig{
			Disk: &StorageDisk{},
		},
	}
}

func (c *Config) Parse() error {
	if err := c.parseFlags(); err != nil {
		return err
	}

	if err := c.parseEnv(); err != nil {
		return err
	}

	return c.validate()
}

func (c *Config) parseFlags() error {
	flag.Parse()

	// set server defaults
	c.LogLevel = *logLevel
	c.WillowPort = *willowPort
	c.MetricsPort = *metricsPort

	// set queue defaults
	c.QueueConfig.MaxSize = *queueMaxSize
	c.QueueConfig.DefaultSize = *queueMaxSize
	c.QueueConfig.DeadLetterMaxSize = *deadLetterQueueMaxSize

	// set storage type
	switch *storageType {
	case string(DiskStorage):
		c.StorageConfig.Type = DiskStorage
		c.StorageConfig.Disk.StorageDir = *diskStorageDir
	case string(MemoryStorage):
		c.StorageConfig.Type = MemoryStorage
	default:
	}

	return nil
}

func (c *Config) parseEnv() error {
	// set server defaults
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
	if willowPort := os.Getenv("WILLOW_PORT"); willowPort != "" {
		c.WillowPort = willowPort
	}
	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		c.MetricsPort = metricsPort
	}

	// set queue defaults
	if queueMaxSize := os.Getenv("QUEUE_MAX_SIZE"); queueMaxSize != "" {
		maxSize, err := strconv.ParseUint(queueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse QUEUE_MAX_SIZE: %w", err)
		}
		c.QueueConfig.MaxSize = maxSize
	}
	if queueMaxSize := os.Getenv("QUEUE_DEFAULT_SIZE"); queueMaxSize != "" {
		maxSize, err := strconv.ParseUint(queueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse QUEUE_DEFAULT_SIZE: %w", err)
		}
		c.QueueConfig.MaxSize = maxSize
	}
	if deadLetterQueueMaxSize := os.Getenv("DEAD_LETTER_QUEUE_MAX_SIZE"); deadLetterQueueMaxSize != "" {
		maxSize, err := strconv.ParseUint(deadLetterQueueMaxSize, 10, 64)
		if err != nil {
			return fmt.Errorf("Failed to parse DEAD_LETTER_QUEUE_MAX_SIZE: %w", err)
		}
		c.QueueConfig.DeadLetterMaxSize = maxSize
	}

	// storage config
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		switch storageType {
		case string(DiskStorage):
			c.StorageConfig.Type = DiskStorage

			if diskStorageLocation := os.Getenv("DISK_STORAGE_DIR"); diskStorageLocation != "" {
				c.StorageConfig.Disk.StorageDir = diskStorageLocation
			}
		case string(MemoryStorage):
			c.StorageConfig.Type = MemoryStorage
		default:
		}
	}

	return nil
}

func (c *Config) validate() error {
	if !(c.LogLevel == "debug" || c.LogLevel == "info") {
		return fmt.Errorf("Expected config 'LogLevel' to be [debug | info]. Received: '%s'", c.LogLevel)
	}

	switch c.StorageConfig.Type {
	case DiskStorage:
		if c.StorageConfig.Disk.StorageDir == "" {
			return fmt.Errorf("'disk-storage-dir' is required when storage type is 'disk'")
		}
	case MemoryStorage:
		// nothing to do here
	default:
		return fmt.Errorf("invalid storage type selected: %s", *storageType)
	}

	return nil
}
