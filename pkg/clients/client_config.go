package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/DanLavine/willow/pkg/models/api"
)

type Config struct {
	// remote url address
	// I.E: https://selfdeployed:8080
	URL string

	// Content type for the client to be using when sending and recieving requests.
	// The service will always respond with the content type sent by the client. The main
	// thought behind this is that different content types can help with various workflows:
	//	1. application/json - easy to understand and reason about when manually testing
	//  2. application/octet-stream - faster to use if there are no rules, and Willow is processing lots of data (not yet implemented)
	ContentType api.ContentType

	// if these values are set, then the config will validate and use the custom tls keys file for https.
	// All of these should be the absolute path to the files
	//
	// custom ca cert if generated with self signed certificates
	CAFile string
	// client key to match server's
	ClientKeyFile string
	// client crt to match server's
	ClientCRTFile string

	// parsed out root ca
	rootCAs *x509.CertPool
	// parsed out CA certs
	cert tls.Certificate
}

func (cfg *Config) Validate() error {
	if cfg.URL == "" {
		return fmt.Errorf("client's Config.URL cannot be empty")
	}

	switch cfg.ContentType {
	case api.ContentTypeJSON:
		// these are all valid
	default:
		return fmt.Errorf("unknown content type: %s", cfg.ContentType)
	}

	if cfg.CAFile == "" && cfg.ClientCRTFile == "" && cfg.ClientKeyFile == "" {
		// this is fine, nothing to do since they are all nil
	} else {
		// ensure all provided certs are here
		if cfg.CAFile == "" || cfg.ClientCRTFile == "" || cfg.ClientKeyFile == "" {
			return fmt.Errorf("when providing custom certs, all 3 values must be provided [CAFile | ClientKeyFile | ClienCRTFile]")
		}

		// parse root ca
		rootCAData, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return fmt.Errorf("failed to read the CAFile: %w", err)
		}

		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM([]byte(rootCAData)); !ok {
			return fmt.Errorf("error parsing CAFile")
		}

		// parse client certs
		cert, err := tls.LoadX509KeyPair(cfg.ClientCRTFile, cfg.ClientKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read the ClientKeyFile or ClientCRTFile: %w", err)
		}

		cfg.rootCAs = rootCAs
		cfg.cert = cert
	}

	return nil
}

func (cfg *Config) RootCAs() *x509.CertPool {
	return cfg.rootCAs
}

func (cfg *Config) Cert() tls.Certificate {
	return cfg.cert
}
