package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/DanLavine/urlrouter"
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
	mux := urlrouter.New()

	// health api
	mux.HandleFunc("GET", "/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// OpenAPI endpoints
	// api url to server all the OpenAPI files
	_, currentDir, _, _ := runtime.Caller(0)
	mux.HandleFunc("GET", "/docs/openapi/", func(w http.ResponseWriter, r *http.Request) {
		handle := http.StripPrefix("/docs/openapi/", http.FileServer(http.Dir(filepath.Join(currentDir, "..", "..", "..", "docs", "openapi"))))
		handle.ServeHTTP(w, r)
	})

	// ui that calls the files and knows what to do
	mux.HandleFunc("GET", "/docs", func(w http.ResponseWriter, r *http.Request) {
		middleware.Redoc(middleware.RedocOpts{
			BasePath: "/",
			Path:     "docs",
			SpecURL:  "/docs/openapi/locker/openapi.yaml",
			Title:    "Locker API Documentation",
		}, nil).ServeHTTP(w, r)
	})

	// crud operations for group rules
	mux.HandleFunc("POST", "/v1/locks", locker.v1Handler.Create)
	mux.HandleFunc("DELETE", "/v1/locks/:_associated_id", locker.v1Handler.Delete)
	mux.HandleFunc("POST", "/v1/locks/:_associated_id/heartbeat", locker.v1Handler.Heartbeat)

	// Admin APIs
	// TODO: Need to actual account for auth for this
	mux.HandleFunc("GET", "/v1/locks", locker.v1Handler.List)

	locker.server.Handler = mux

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownErr <- locker.server.Shutdown(context.Background())
	}()

	// return any error other than the server closed
	logger.Info("Locker TCP server running", zap.String("port", *locker.config.LockerPort))
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
