package config

import (
	"flag"
	"fmt"
	"os"
)

var (
	logLevel    = flag.String("log-level", "info", "log level [debug | info]. Can be set by env var LOG_LEVEL")
	willowPort  = flag.String("willow-port", "8080", "willow server port. Can be set by env var WILLOW_PORT")
	metricsPort = flag.String("metrics-port", "8081", "willow server metrics port. can be set by env var METRICS_PORT")
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

	// Type of storage we are using
	StorageType StorageType

	// Disk Storage Configuration
	// Valid fields: [disk]
	DiskStorageDir string
}

func Default() *Config {
	return &Config{}
}

func (c *Config) Parse() error {
	c.parseFlags()
	c.parseEnv()

	return c.validate()
}

func (c *Config) parseEnv() {
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

	if diskStorageLocation := os.Getenv("DISK_STORAGE_DIR"); diskStorageLocation != "" {
		c.DiskStorageDir = diskStorageLocation
	}
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

	if diskStorageDir != nil {
		c.DiskStorageDir = *diskStorageDir
	}
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
