package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/DanLavine/willow/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

//go:generate mockgen -destination=serverfakes/lmite_handler_mock.go -package=serverfakes github.com/DanLavine/willow/internal/server LimiterHandler
type LimiterHandler interface {
	// group rule operations
	CreateGroupRule(w http.ResponseWriter, r *http.Request)
	FindGroupRule(w http.ResponseWriter, r *http.Request)
	UpdateGroupRule(w http.ResponseWriter, r *http.Request)
	DeleteGroupRule(w http.ResponseWriter, r *http.Request)

	// item operations
	IncrementItem(w http.ResponseWriter, r *http.Request)
	DecrementItem(w http.ResponseWriter, r *http.Request)

	// client operations?
}

type limiterTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LimiterConfig
	server *http.Server

	limiterHandler LimiterHandler
}

func NewLimiterTCP(logger *zap.Logger, config *config.LimiterConfig, limiterHandler LimiterHandler) *limiterTCP {
	return &limiterTCP{
		closed:         false,
		logger:         logger.Named("limiterTCP_server"),
		config:         config,
		limiterHandler: limiterHandler,
	}
}

func (limiter *limiterTCP) Cleanup() error { return nil }
func (limiter *limiterTCP) Initialize() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", *limiter.config.LimiterPort),
	}

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

	// certs loaded successfully
	return nil
}

func (limiter *limiterTCP) Execute(ctx context.Context) error {
	logger := limiter.logger

	// capture any errors from the server
	errChan := make(chan error, 1)
	defer close(errChan)

	mux := http.NewServeMux()

	// crud operations for group rules
	// These operations seem more like a normal DB that I want to do...
	mux.HandleFunc("/v1/group_rules/create", limiter.limiterHandler.CreateGroupRule)
	mux.HandleFunc("/v1/group_rules/find", limiter.limiterHandler.FindGroupRule)
	mux.HandleFunc("/v1/group_rules/update", limiter.limiterHandler.UpdateGroupRule)
	mux.HandleFunc("/v1/group_rules/delete", limiter.limiterHandler.DeleteGroupRule)

	// operations to check items against arbitrary rules
	mux.HandleFunc("/v1/items/increment", limiter.limiterHandler.IncrementItem)
	mux.HandleFunc("/v1/items/decrement", limiter.limiterHandler.DecrementItem)
	// delete all items, or a collection of gruped tags?
	// seems like something I would want to do...
	//mux.HandleFunc("/v1/items/delete", nil)

	// client operations?
	// want to kick a bunch of operations that might be pending?
	//mux.HandleFunc("/v1/items/kick", nil)

	// set the server mux
	limiter.server.Handler = mux
	http2.ConfigureServer(limiter.server, &http2.Server{})

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-ctx.Done()
		shutdownErr <- limiter.server.Shutdown(context.Background())
	}()

	// return any error other than the server closed
	logger.Info("Limiter TCP server running")
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

	// wait for any pending connections to drain
	if err := <-shutdownErr; err != nil {
		logger.Error("server shutdown with error", zap.Error(err))
		return err
	}

	logger.Info("server shutdown successfully")
	return nil
}
