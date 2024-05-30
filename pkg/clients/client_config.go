package clients

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type Config struct {
	// remote url address
	// I.E: https://selfdeployed:8080
	URL string

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

	if cfg.CAFile == "" && cfg.ClientCRTFile == "" && cfg.ClientKeyFile == "" {
		// this is fine, nothing to do since they are all nil
	} else if cfg.CAFile != "" && (cfg.ClientCRTFile == "" || cfg.ClientKeyFile == "") {
		return fmt.Errorf("when using custom certs and the 'CAFile' is provided then 'ClientKeyFile' and 'ClienCRTFile' must also be provided")
	} else {
		// ensure all other certs are valid
		if cfg.ClientCRTFile == "" || cfg.ClientKeyFile == "" {
			return fmt.Errorf("when providing custom certs, the key and crt values must be provided [ClientKeyFile | ClienCRTFile]")
		}

		// parse client certs
		cert, err := tls.LoadX509KeyPair(cfg.ClientCRTFile, cfg.ClientKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read the ClientKeyFile or ClientCRTFile: %w", err)
		}

		// optional parse root ca
		if cfg.CAFile != "" {
			rootCAData, err := os.ReadFile(cfg.CAFile)
			if err != nil {
				return fmt.Errorf("failed to read the CAFile: %w", err)
			}

			rootCAs := x509.NewCertPool()
			if ok := rootCAs.AppendCertsFromPEM([]byte(rootCAData)); !ok {
				return fmt.Errorf("error parsing CAFile")
			}

			// validate that the key is corret
			keyData, err := os.ReadFile(cfg.ClientCRTFile)
			if err != nil {
				return fmt.Errorf("failed to read ClientCRTFile: %w", err)
			}

			block, _ := pem.Decode(keyData)
			if block == nil {
				return fmt.Errorf("failed to decode ClientCRTFile")
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("faled to parse ClientCRTFile: %w", err)
			}

			if _, err = cert.Verify(x509.VerifyOptions{Roots: rootCAs}); err != nil {
				return fmt.Errorf("failed to verify certs: %w", err)
			}

			cfg.rootCAs = rootCAs
		}

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
