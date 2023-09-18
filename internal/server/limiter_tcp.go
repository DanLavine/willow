package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/DanLavine/willow/internal/server/versions/v1limiter"
	"github.com/DanLavine/willow/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type limiterTCP struct {
	closed bool

	logger *zap.Logger
	config *config.LimiterConfig
	server *http.Server

	v1ruleHandler v1limiter.LimitRuleHandler
}

func NewLimiterTCP(logger *zap.Logger, config *config.LimiterConfig, v1ruleHandler v1limiter.LimitRuleHandler) *limiterTCP {
	return &limiterTCP{
		closed:        false,
		logger:        logger.Named("limiterTCP_server"),
		config:        config,
		v1ruleHandler: v1ruleHandler,
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
	mux.HandleFunc("/v1/group_rules/create", limiter.v1ruleHandler.Create)
	mux.HandleFunc("/v1/group_rules/set_override", limiter.v1ruleHandler.SetOverride)

	mux.HandleFunc("/v1/group_rules/find", limiter.v1ruleHandler.Find)
	mux.HandleFunc("/v1/group_rules/update", limiter.v1ruleHandler.Update)
	mux.HandleFunc("/v1/group_rules/delete", limiter.v1ruleHandler.Delete)

	// operations to check items against arbitrary rules
	mux.HandleFunc("/v1/items/increment", limiter.v1ruleHandler.Increment)
	mux.HandleFunc("/v1/items/decrement", limiter.v1ruleHandler.Decrement)
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
