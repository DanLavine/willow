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
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type limiterTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LimiterConfig
	server *http.Server

	mux *urlrouter.Router
}

func NewLimiterTCP(logger *zap.Logger, config *config.LimiterConfig, mux *urlrouter.Router) *limiterTCP {
	return &limiterTCP{
		closed: false,
		logger: logger.Named("limiterTCP_server"),
		config: config,
		mux:    mux,
	}
}

func (limiter *limiterTCP) Cleanup() error { return nil }
func (limiter *limiterTCP) Initialize() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", *limiter.config.LimiterPort),
	}

	if !*limiter.config.LimiterInsecureHTTP {
		// load the server CRT and Key
		cert, err := tls.LoadX509KeyPair(*limiter.config.LimiterServerCRT, *limiter.config.LimiterServerKey)
		if err != nil {
			return err
		}

		// add them to the server
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		// load the CA cert if it exists
		if *limiter.config.LimiterCA != "" {
			CaPEM, err := os.ReadFile(*limiter.config.LimiterCA)
			if err != nil {
				return err
			}
			CAs := x509.NewCertPool()
			if !CAs.AppendCertsFromPEM(CaPEM) {
				return fmt.Errorf("failed to parse LimiterCA")
			}

			server.TLSConfig.RootCAs = CAs
		}

		// enforce using http2
		limiter.server = server
		http2.ConfigureServer(limiter.server, &http2.Server{})
	} else {
		// set to http
		limiter.server = server
	}

	// certs loaded successfully
	return nil
}

func (limiter *limiterTCP) Execute(ctx context.Context) error {
	logger := limiter.logger

	// health api doesn't have a version associated with it
	limiter.mux.HandleFunc("GET", "/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// set the server mux
	limiter.server.Handler = limiter.mux

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownErr <- limiter.server.Shutdown(context.Background())
	}()

	// return any error other than the server closed
	logger.Info("Limiter TCP server running", zap.String("port", *limiter.config.LimiterPort))
	if *limiter.config.LimiterInsecureHTTP {
		if err := limiter.server.ListenAndServe(); err != nil {
			select {
			case <-ctx.Done():
				if err != http.ErrServerClosed {
					// must be an unexpected error during shutdown
					return err
				}

				// context was closed and server closed error. clean shutdown case
			default:
				// always return the error if the context was not closed
				return err
			}
		}
	} else {
		if err := limiter.server.ListenAndServeTLS("", ""); err != nil {
			select {
			case <-ctx.Done():
				if err != http.ErrServerClosed {
					// must be an unexpected error during shutdown
					return err
				}

				// context was closed and server closed error. clean shutdown case
			default:
				// always return the error if the context was not closed
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
