package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type LockerTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LockerConfig
	server *http.Server

	mux *urlrouter.Router

	connTracker btree.BTree
}

func NewLockerTCP(logger *zap.Logger, config *config.LockerConfig, mux *urlrouter.Router) *LockerTCP {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &LockerTCP{
		closed:      false,
		logger:      logger.Named("LockerTCP_server"),
		config:      config,
		mux:         mux,
		connTracker: tree,
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

	// health api doesn't have a version associated with it
	locker.mux.HandleFunc("GET", "/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	locker.server.Handler = locker.mux

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
