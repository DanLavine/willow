package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/server/versions/v1server"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type LockerTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LockerConfig
	server *http.Server

	v1Handler v1server.LockerHandler
}

func NewLockerTCP(logger *zap.Logger, config *config.LockerConfig, v1Handler v1server.LockerHandler) *LockerTCP {
	return &LockerTCP{
		closed:    false,
		logger:    logger.Named("LockerTCP_server"),
		config:    config,
		v1Handler: v1Handler,
	}
}

func (locker *LockerTCP) Cleanup() error { return nil }
func (locker *LockerTCP) Initialize() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", *locker.config.LockerPort),
	}

	// load the server CRT and Key
	cert, err := tls.LoadX509KeyPair(*locker.config.LockerServerCRT, *locker.config.LockerServerKey)
	if err != nil {
		return err
	}

	// add them to the server
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// load the CA cert if it exists
	if *locker.config.LockerCA != "" {
		CaPEM, err := os.ReadFile(*locker.config.LockerCA)
		if err != nil {
			return err
		}
		CAs := x509.NewCertPool()
		if !CAs.AppendCertsFromPEM(CaPEM) {
			return fmt.Errorf("failed to parse LockerCA")
		}

		server.TLSConfig.RootCAs = CAs
	}

	// enforce using http2
	locker.server = server

	// certs loaded successfully
	return nil
}

func (Locker *LockerTCP) Execute(ctx context.Context) error {
	logger := Locker.logger

	// capture any errors from the server
	errChan := make(chan error, 1)
	defer close(errChan)

	mux := http.NewServeMux()

	// crud operations for group rules
	// These operations seem more like a normal DB that I want to do...
	mux.HandleFunc("/v1/locker/create", Locker.v1Handler.Create)
	mux.HandleFunc("/v1/locker/list", Locker.v1Handler.List)
	mux.HandleFunc("/v1/locker/list", Locker.v1Handler.Delete)

	// set the server mux
	Locker.server.Handler = mux
	http2.ConfigureServer(Locker.server, &http2.Server{})

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownErr <- Locker.server.Shutdown(context.Background())
	}()
	// return any error other than the server closed
	logger.Info("Locker TCP server running")
	if err := Locker.server.ListenAndServeTLS("", ""); err != nil {
		select {
		case <-ctx.Done():
			if err != http.ErrServerClosed {
				// must be an unexpected error during shutdown
				return err
			}

			// context was closed and server closed error. clean shutdown case
		default:
			// always return the error if the parent context was not closed
			return err
		}
	}

	// wait for any pending connections to drain
	if err := <-shutdownErr; err != nil {
		logger.Error("server shutdown with error", zap.Error(err))
		return err
	}

	logger.Info("server shutdown successfully")
	return nil
}
