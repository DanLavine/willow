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
	queueMaxEnries           = flag.Int("queue-max-entries", 4096, "max entries that can be enqueued at once. This includes any entries that need to be retried. Can be set by env var QUEUE_MAX_ENTRIES")
	deadLetterQueueMaxEnries = flag.Int("dead-letter-queue-max-entries", 100, "max entries that can be stored in the dead letter queue. Can be set by env var DEAD_LETTER_QUEUE_MAX_ENTRIES")

	// storage configurations
	storageType = flag.String("storage-type", "disk", "storage type to use for persistence [disk]. Can be set by env var STORAGE_TYPE")
	// disck storage configurations
	diskStorageDir = flag.String("disk-storage-dir", "", "root location on disk where to save storage data. Can be set by env var DISK_STORAGE_DIR")
)

type StorageType string

const (
	DiskStorage    StorageType = "disk"
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
	ConfigQueue ConfigQueue
}

// Queue configuration
type ConfigQueue struct {
	// max number of entries any queue can be configured for
	QueueMaxEntries           int
	DeadLetterQueueMaxEntries int

	// Type of storage we are using
	StorageType StorageType
	ConfigDisk  *ConfigDisk
}

// Disk Storage Configuration
type ConfigDisk struct {
	// Valid fields: [disk]
	StorageDir string
}

func Default() *Config {
	return &Config{}
}

func (c *Config) Parse() error {
	c.parseFlags()

	if err := c.parseEnv(); err != nil {
		return err
	}

	return nil
}

func (c *Config) parseFlags() {
	flag.Parse()

	c.LogLevel = *logLevel
	c.WillowPort = *willowPort
	c.MetricsPort = *metricsPort

	// set queue configurations
	if queueMaxEnries != nil {
		c.ConfigQueue.QueueMaxEntries = *queueMaxEnries
	}
	if deadLetterQueueMaxEnries != nil {
		c.ConfigQueue.DeadLetterQueueMaxEntries = *deadLetterQueueMaxEnries
	}

	// set storage type
	if storageType == nil {
		c.ConfigQueue.StorageType = InvalidStorage
	} else {
		switch *storageType {
		case "disk":
			c.ConfigQueue.StorageType = DiskStorage
			c.ConfigQueue.ConfigDisk = &ConfigDisk{}

			if diskStorageDir != nil {
				c.ConfigQueue.ConfigDisk.StorageDir = *diskStorageDir
			}
		default:
			c.ConfigQueue.StorageType = InvalidStorage
		}
	}
}

func (c *Config) parseEnv() error {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}

	if willowPort := os.Getenv("WILLOW_PORT"); willowPort != "" {
		c.WillowPort = willowPort
	}

	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		c.MetricsPort = metricsPort
	}

	//general queue config
	if queueMaxEntries := os.Getenv("QUEUE_MAX_ENTRIES"); queueMaxEntries != "" {
		maxEntries, err := strconv.Atoi(queueMaxEntries)
		if err != nil {
			return fmt.Errorf("Failed to parse QUEUE_MAX_ENTRIES: %w", err)
		}

		c.ConfigQueue.QueueMaxEntries = maxEntries
	}
	if deadLetterQueueMaxEntries := os.Getenv("DEAD_LETTER_QUEUE_MAX_ENTRIES"); deadLetterQueueMaxEntries != "" {
		maxEntries, err := strconv.Atoi(deadLetterQueueMaxEntries)
		if err != nil {
			return fmt.Errorf("Failed to parse DEAD_LETTER_QUEUE_MAX_ENTRIES: %w", err)
		}

		c.ConfigQueue.DeadLetterQueueMaxEntries = maxEntries
	}

	// storage config
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		switch storageType {
		case "disk":
			c.ConfigQueue.StorageType = DiskStorage
			if c.ConfigQueue.ConfigDisk == nil {
				c.ConfigQueue.ConfigDisk = &ConfigDisk{}
			}

			if diskStorageLocation := os.Getenv("DISK_STORAGE_DIR"); diskStorageLocation != "" {
				c.ConfigQueue.ConfigDisk.StorageDir = diskStorageLocation
			}
		default:
			c.ConfigQueue.StorageType = InvalidStorage
		}
	}

	return nil
}

func (c *Config) Validate() error {
	if !(c.LogLevel == "debug" || c.LogLevel == "info") {
		return fmt.Errorf("Expected config 'LogLevel' to be [debug | info]. Received: '%s'", c.LogLevel)
	}

	return c.ConfigQueue.Validate()
}

func (c ConfigQueue) Validate() error {
	switch storageType := c.StorageType; storageType {
	case "disk":
		if c.ConfigDisk.StorageDir == "" {
			return fmt.Errorf("Expected config 'StorageDir' to be set when 'StorageType = disk'")
		}
	default:
		return fmt.Errorf("Expected config 'StorageType' to be [disk]. Received: '%s'", storageType)
	}

	return nil

}
