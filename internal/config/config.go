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
	storageType = flag.String("storage-type", "disk", "storage type to use for persistence [disk]. Can be set by env var STORAGE_TYPE")

	// general storage configurations
	queueMaxEnries = flag.Int("queue-max-entries", 4096, "max entries that can be enqueued at once. This includes any entries that need to be retried. Can be set by env var QUEUE_MAX_ENTRIES")

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

	// Type of storage we are using
	StorageType StorageType

	// general queue configuration
	QueueMaxEntries int

	// Disk Storage Configuration
	// Valid fields: [disk]
	DiskStorageDir string
}

func Default() *Config {
	return &Config{}
}

func (c *Config) Parse() error {
	c.parseFlags()

	if err := c.parseEnv(); err != nil {
		return err
	}

	return c.validate()
}

func (c *Config) parseFlags() {
	flag.Parse()

	c.LogLevel = *logLevel
	c.WillowPort = *willowPort
	c.MetricsPort = *metricsPort

	if storageType == nil {
		c.StorageType = InvalidStorage
	} else {
		switch *storageType {
		case "disk":
			c.StorageType = DiskStorage
		default:
			c.StorageType = InvalidStorage
		}
	}

	c.QueueMaxEntries = *queueMaxEnries

	if diskStorageDir != nil {
		c.DiskStorageDir = *diskStorageDir
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

	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		switch storageType {
		case "disk":
			c.StorageType = DiskStorage
		default:
			c.StorageType = InvalidStorage
		}
	}

	if queueMaxEntries := os.Getenv("QUEUE_MAX_ENTRIES"); queueMaxEntries != "" {
		maxEntries, err := strconv.Atoi(queueMaxEntries)
		if err != nil {
			return fmt.Errorf("Failed to parse QUEUE_MAX_ENTRIES: %w", err)
		}

		c.QueueMaxEntries = maxEntries
	}

	if diskStorageLocation := os.Getenv("DISK_STORAGE_DIR"); diskStorageLocation != "" {
		c.DiskStorageDir = diskStorageLocation
	}

	return nil
}

func (c *Config) validate() error {
	if !(c.LogLevel == "debug" || c.LogLevel == "info") {
		return fmt.Errorf("Expected config 'LogLevel' to be [debug | info]. Received: '%s'", c.LogLevel)
	}

	switch storageType := c.StorageType; storageType {
	case "disk":
		if c.DiskStorageDir == "" {
			return fmt.Errorf("Expected config 'DiskStorageDir' to be set when 'StorageType = disk'")
		}
	default:
		return fmt.Errorf("Expected config 'StorageType' to be [disk]. Received: '%s'", storageType)
	}

	return nil
}
