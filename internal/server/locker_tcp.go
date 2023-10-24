package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/server/versions/v1server"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type LockerTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LockerConfig
	server *http.Server

	connTracker btree.BTree

	v1Handler v1server.LockerHandler
}

func NewLockerTCP(logger *zap.Logger, config *config.LockerConfig, v1Handler v1server.LockerHandler) *LockerTCP {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &LockerTCP{
		closed:      false,
		logger:      logger.Named("LockerTCP_server"),
		config:      config,
		connTracker: tree,
		v1Handler:   v1Handler,
	}
}

func (locker *LockerTCP) Cleanup() error { return nil }
func (locker *LockerTCP) Initialize() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", *locker.config.LockerPort),
	}

	if !*locker.config.LockerInsecureHTTP {
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
		http2.ConfigureServer(locker.server, &http2.Server{})
	} else {
		// set to default http server
		locker.server = server
	}

	// certs loaded successfully
	return nil
}

func (locker *LockerTCP) Execute(ctx context.Context) error {
	logger := locker.logger

	// capture any errors from the server
	errChan := make(chan error, 1)
	defer close(errChan)

	mux := http.NewServeMux()

	// OpenAPI endpoints
	// api url to server all the OpenAPI files
	mux.Handle("/docs/openapi/", http.StripPrefix("/docs/openapi/", http.FileServer(http.Dir("./docs/openapi/locker"))))

	// ui that calls the files and knows what to do
	mux.Handle("/docs", middleware.Redoc(middleware.RedocOpts{
		BasePath: "/",
		Path:     "docs",
		SpecURL:  "/docs/openapi/openapi.yaml",
		Title:    "Locker API Documentation",
	}, nil))

	// crud operations for group rules
	// These operations seem more like a normal DB that I want to do...
	mux.HandleFunc("/v1/locker/create", locker.v1Handler.Create) // pass the ctx here so clients can clean up when the server shutsdown
	mux.HandleFunc("/v1/locker/heartbeat", locker.v1Handler.Heartbeat)
	mux.HandleFunc("/v1/locker/list", locker.v1Handler.List)
	mux.HandleFunc("/v1/locker/delete", locker.v1Handler.Delete)

	locker.server.Handler = mux

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownErr <- locker.server.Shutdown(context.Background())
	}()

	// return any error other than the server closed
	logger.Info("Locker TCP server running")
	if *locker.config.LockerInsecureHTTP {
		if err := locker.server.ListenAndServe(); err != nil {
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
	} else {
		if err := locker.server.ListenAndServeTLS("", ""); err != nil {
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
	}

	// wait for any pending connections to drain
	if err := <-shutdownErr; err != nil {
		logger.Error("server shutdown with error", zap.Error(err))
		return err
	}

	logger.Info("server shutdown successfully")
	return nil
}
