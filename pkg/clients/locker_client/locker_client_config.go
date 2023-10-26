package lockerclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type Config struct {
	// remote url address
	// I.E: https://selfdeployedlocker/
	URL string

	// if these values are set, then the config will validate and use the custom tls keys file for https.
	// All of these should be the absolute path to the file
	//
	// custom ca cert if generated with self signed certificates
	LockerCAFile string
	// client key to match server's
	LockerClientKeyFile string
	// client crt to match server's
	LockerClientCRTFile string

	// parsed out root ca
	rootCAs *x509.CertPool
	// parsed out CA certs
	cert tls.Certificate
}

func (cfg *Config) Vaidate() error {
	if cfg.URL == "" {
		return fmt.Errorf("LockerClient's Config.URL cannot be empty")
	}

	if cfg.LockerCAFile == "" && cfg.LockerClientCRTFile == "" && cfg.LockerClientKeyFile == "" {
		// this is fine, nothing to do since they are all nil
	} else {
		// ensure all provided certs are here
		if cfg.LockerCAFile == "" || cfg.LockerClientCRTFile == "" || cfg.LockerClientKeyFile == "" {
			return fmt.Errorf("when providing custom certs, all 3 values must be provided [LockerCAFile | LockerClientKeyFile | LockerClienCRTFile]")
		}

		// parse root ca
		rootCAData, err := os.ReadFile(cfg.LockerCAFile)
		if err != nil {
			return fmt.Errorf("failed to read the LockerCAFile: %w", err)
		}

		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM([]byte(rootCAData)); !ok {
			return fmt.Errorf("error parsing LockerCAFile")
		}

		// parse client certs
		cert, err := tls.LoadX509KeyPair(cfg.LockerClientCRTFile, cfg.LockerClientKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read the LockerClientKeyFile or LockerClientCRTFile: %w", err)
		}

		cfg.rootCAs = rootCAs
		cfg.cert = cert
	}

	return nil
}
