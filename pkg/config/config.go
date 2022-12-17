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
)

type Config struct {
	// Log Level [debug | info]
	LogLevel string

	// port to run the willow tcp server
	WillowPort string

	// port to run the metrics http server on
	MetricsPort string
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
}

func (c *Config) parseFlags() {
	flag.Parse()

	c.LogLevel = *logLevel
	c.WillowPort = *willowPort
	c.MetricsPort = *metricsPort
}

func (c *Config) validate() error {
	if !(c.LogLevel == "debug" || c.LogLevel == "info") {
		return fmt.Errorf("Expected config 'LogLevel' to be either [debug | info]. Received: '%s'", c.LogLevel)
	}

	return nil
}
